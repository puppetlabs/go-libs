package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

//N.B. The rate limiting library does not play nicely with httptest hence rate limiting is not tested.
var (
	testEndpoint      = "/helloworld"
	readinessEndpoint = "/readiness"
	metricsEndpoint   = "/metrics"
	allowedOrigin     = "https://wwww.puppet.com/"
)

type readinessResp struct {
	Status string
}

type headers struct {
	Name  string
	Value string
}

var validMethods = []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete}

func setupService(cfg *Config) (*Service, error) {
	svc, err := NewService(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create service config %s", err)
	}
	return svc, nil
}

func sendRequest(svc *Service, method string, url string, reqHeaders ...headers) (*httptest.ResponseRecorder, error) {
	rr := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create  client request due to error %s", err)
	}

	for _, header := range reqHeaders {
		req.Header.Set(header.Name, header.Value)
	}

	svc.Handler.ServeHTTP(rr, req)
	return rr, nil
}

func checkResponseCode(method string, url string, cfg Config, code int, reqHeaders ...headers) (*httptest.ResponseRecorder, error) {

	svc, err := setupService(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create service config %s", err)
	}

	rr, err := sendRequest(svc, method, url, reqHeaders...)
	if err != nil {
		return nil, fmt.Errorf("unable to create  client request due to error %s", err)
	}
	if rr.Code != code {
		return nil, fmt.Errorf("Unexpected response code %d. Expected %d", rr.Code, code)
	}

	return rr, nil
}

func checkMultipleResponseCodes(methods []string, url string, cfg Config, code int, reqHeaders ...headers) ([]*httptest.ResponseRecorder, error) {

	svc, err := setupService(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create service config %s", err)
	}

	var responses []*httptest.ResponseRecorder

	for _, method := range methods {
		rr, err := sendRequest(svc, method, url, reqHeaders...)
		if err != nil {
			return nil, fmt.Errorf("unable to create  client request due to error %s", err)
		}
		if rr.Code != code {
			return nil, fmt.Errorf("Unexpected response code %d. Expected %d", rr.Code, code)
		}

		responses = append(responses, rr)
	}

	return responses, nil
}

//helloWorldHandler is a placeholder
func helloWorldHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World.")
	}
}

func handlerWithWait(waitLength time.Duration) func(c *gin.Context) {
	time.Sleep(waitLength * time.Second)
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World.")
	}
}

//returnWithResponseCode returns the code that is passed in
func returnWithResponseCode(httpStatus int) func(c *gin.Context) {
	return func(c *gin.Context) {
		//Return what would be returned if access is denied.
		c.AbortWithStatus(httpStatus)
	}
}

func TestNoListenAddressErrors(t *testing.T) {
	cfg := Config{
		Handlers: []Handler{{Method: AnyMethod, Handler: helloWorldHandler(), Path: "dummy"}},
	}

	_, err := NewService(&cfg)
	if err == nil {
		t.Error("No listen address should cause error.")
	}
}

func TestNoHandlerErrors(t *testing.T) {
	cfg := Config{
		ListenAddress: ":8888",
	}

	_, err := NewService(&cfg)
	if err == nil {
		t.Error("No registered handlers should cause error.")
	}
}

func TestAnyMethod(t *testing.T) {
	cfg := Config{
		ListenAddress: ":8888",
		Handlers:      []Handler{{Method: AnyMethod, Handler: helloWorldHandler(), Path: testEndpoint}},
	}

	_, err := checkMultipleResponseCodes(validMethods, testEndpoint, cfg, http.StatusOK)
	if err != nil {
		t.Error(err)
	}
}

func TestIndividualHTTPMethodsGoodEnpoints(t *testing.T) {
	cfg := Config{
		ListenAddress: ":8888",
		Handlers: []Handler{
			{Method: http.MethodGet, Handler: helloWorldHandler(), Path: testEndpoint},
			{Method: http.MethodPost, Handler: helloWorldHandler(), Path: testEndpoint},
			{Method: http.MethodPatch, Handler: helloWorldHandler(), Path: testEndpoint},
			{Method: http.MethodDelete, Handler: helloWorldHandler(), Path: testEndpoint},
		},
	}

	_, err := checkMultipleResponseCodes(validMethods, testEndpoint, cfg, http.StatusOK)
	if err != nil {
		t.Error("Good endpoints should be returning ok status codes")
	}
}

func TestIndividualHTTPMethodsBadEndpoints(t *testing.T) {
	cfg := Config{
		ListenAddress: ":8888",
		Handlers: []Handler{
			{Method: http.MethodGet, Handler: helloWorldHandler(), Path: "/valid"},
		},
	}

	_, err := checkMultipleResponseCodes(validMethods, "/notvalid", cfg, http.StatusNotFound)
	if err != nil {
		t.Error("Bad endpoints should be returning not found status codes")
	}
}

func TestClashingHandlers(t *testing.T) {
	cfg := Config{
		ListenAddress: ":8888",
		Handlers: []Handler{
			{Method: http.MethodGet, Handler: helloWorldHandler(), Path: testEndpoint},
			{Method: http.MethodGet, Handler: helloWorldHandler(), Path: testEndpoint},
		},
	}

	_, err := NewService(&cfg)
	if err == nil {
		t.Error("Clashing handlers should cause error")
	}
}

// func TestReadTimeout(t *testing.T) {
// 	cfg := Config{
// 		ListenAddress: ":8888",
// 		Handlers: []Handler{
// 			{Method: http.MethodGet, Handler: handlerWithWait(2), Path: testEndpoint},
// 		},
// 		ReadTimeout:  1 * time.Microsecond,
// 		WriteTimeout: 1 * time.Microsecond,
// 	}

// 	startTime := time.Now()

// 	_, err := checkResponseCode(http.MethodGet, testEndpoint, cfg, http.StatusOK)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	duration := time.Now().Sub(startTime)

// 	if duration > 1*time.Second {
// 		t.Errorf("Greater: %v", duration)
// 	} else {
// 		t.Errorf("less than: %v", duration)
// 	}

// }

func TestRegisteredHandlerReturnsCorrectResponse(t *testing.T) {
	cfg := Config{
		ListenAddress: ":8888",
		Handlers:      []Handler{{Method: AnyMethod, Handler: helloWorldHandler(), Path: testEndpoint}},
	}
	rr, err := checkResponseCode(http.MethodGet, testEndpoint, cfg, http.StatusOK)
	if err != nil {
		t.Error(err)
	}

	// Check the response body is what we expect.
	expected := "Hello World."
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got <%v> want <%v>",
			rr.Body.String(), expected)
	}

}

func TestMultipleGroupsMiddleware(t *testing.T) {
	mwHandlers := []MiddlewareHandler{{Groups: []string{"returnAccepted"}, Handler: returnWithResponseCode(http.StatusAccepted)},
		{Groups: []string{"returnAlreadyReported"}, Handler: returnWithResponseCode(http.StatusAlreadyReported)}}

	handlers := []Handler{{Method: AnyMethod, Handler: helloWorldHandler(), Path: "/accept", Group: "returnAccepted"},
		{Method: AnyMethod, Handler: helloWorldHandler(), Path: "/reported", Group: "returnAlreadyReported"}}

	cfg := Config{
		ListenAddress:      ":8888",
		Handlers:           handlers,
		MiddlewareHandlers: mwHandlers,
	}

	svc, err := setupService(&cfg)
	if err != nil {
		t.Error(err)
	}

	acceptRr, err := sendRequest(svc, http.MethodGet, "/accept")
	if err != nil {
		t.Error(err)
	}

	if acceptRr.Code != http.StatusAccepted {
		t.Errorf("Expected status %d but got %d.", http.StatusAccepted, acceptRr.Code)
	}

	reportRr, err := sendRequest(svc, http.MethodGet, "/reported")
	if err != nil {
		t.Error(err)
	}

	if reportRr.Code != http.StatusAlreadyReported {
		t.Errorf("Expected status %d but got %d.", http.StatusAlreadyReported, reportRr.Code)
	}
}

func TestMultipleGroupsCors(t *testing.T) {
	cfg := Config{
		ListenAddress: ":8888",
		Handlers: []Handler{{Group: "alloworigin", Method: http.MethodGet, Handler: helloWorldHandler(), Path: testEndpoint},
			{Group: "nocors", Method: http.MethodGet, Handler: helloWorldHandler(), Path: "/nocors"}},
		Cors: &CorsConfig{Groups: []string{"alloworigin"}, Enabled: true, OverrideCfg: &cors.Config{AllowOrigins: []string{allowedOrigin}}},
	}

	originHeader := headers{Name: "Origin", Value: "https://wwww.notallowedorigin.com/"}
	svc, err := setupService(&cfg)
	if err != nil {
		t.Error(err)
	}

	forbiddenRr, err := sendRequest(svc, http.MethodGet, testEndpoint, originHeader)
	if err != nil {
		t.Error(err)
	}

	if forbiddenRr.Code != http.StatusForbidden {
		t.Errorf("Expected status %d but got %d.", http.StatusForbidden, forbiddenRr.Code)
	}

	allowedRr, err := sendRequest(svc, http.MethodGet, "/nocors", originHeader)
	if err != nil {
		t.Error(err)
	}

	if allowedRr.Code != http.StatusOK {
		t.Errorf("Expected status %d but got %d.", http.StatusForbidden, allowedRr.Code)
	}

}

func TestReadinessHandler(t *testing.T) {
	cfg := Config{
		ListenAddress:  ":8888",
		Handlers:       []Handler{{Method: AnyMethod, Handler: helloWorldHandler(), Path: testEndpoint}},
		ReadinessCheck: true,
	}
	rr, err := checkResponseCode(http.MethodGet, readinessEndpoint, cfg, http.StatusOK)
	if err != nil {
		t.Error(err)
	}

	// Check the response body is what we expect.
	var resp readinessResp
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("Unable to unmarshal response %s.", err)
	}
	expected := readinessResp{Status: "UP"}
	if !reflect.DeepEqual(resp, expected) {
		t.Errorf("Status not as expected: got <%v> want <%v>",
			resp, expected)
	}
}

func TestAddMiddleware(t *testing.T) {
	cfg := Config{
		ListenAddress:      ":8888",
		Handlers:           []Handler{{Method: http.MethodGet, Handler: helloWorldHandler(), Path: testEndpoint}},
		MiddlewareHandlers: []MiddlewareHandler{{Handler: returnWithResponseCode(http.StatusContinue)}},
	}
	_, err := checkResponseCode(http.MethodGet, testEndpoint, cfg, http.StatusContinue)
	if err != nil {
		t.Error(err)
	}
}

func TestCorsForbidden(t *testing.T) {
	cfg := Config{
		ListenAddress: ":8888",
		Handlers:      []Handler{{Method: http.MethodGet, Handler: helloWorldHandler(), Path: testEndpoint}},
		Cors:          &CorsConfig{Enabled: true, OverrideCfg: &cors.Config{AllowOrigins: []string{allowedOrigin}}},
	}

	originHeader := headers{Name: "Origin", Value: "https://wwww.notallowedorigin.com/"}
	_, err := checkResponseCode(http.MethodGet, testEndpoint, cfg, http.StatusForbidden, originHeader)
	if err != nil {
		t.Error(err)
	}
}

func TestCorsAllowed(t *testing.T) {
	cfg := Config{
		ListenAddress: ":8888",
		Handlers:      []Handler{{Method: http.MethodGet, Handler: helloWorldHandler(), Path: testEndpoint}},
		Cors:          &CorsConfig{Enabled: true, OverrideCfg: &cors.Config{AllowOrigins: []string{allowedOrigin}}},
	}

	originHeader := headers{Name: "Origin", Value: allowedOrigin}
	_, err := checkResponseCode(http.MethodGet, testEndpoint, cfg, http.StatusOK, originHeader)
	if err != nil {
		t.Error(err)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	cfg := Config{
		ListenAddress: ":8888",
		Handlers:      []Handler{{Method: http.MethodGet, Handler: helloWorldHandler(), Path: testEndpoint}},
		Metrics:       true,
	}

	_, err := checkResponseCode(http.MethodGet, metricsEndpoint, cfg, http.StatusOK)
	if err != nil {
		t.Error(err)
	}
}
