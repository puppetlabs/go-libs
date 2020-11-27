package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

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

func checkResponseCode(method string, url string, cfg Config, code int, reqHeaders ...headers) (*httptest.ResponseRecorder, error) {
	rr := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create  client request due to error %s", err)
	}

	for _, header := range reqHeaders {
		req.Header.Set(header.Name, header.Value)
	}

	svc, err := NewService(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create service config %s", err)
	}
	svc.Handler.ServeHTTP(rr, req)

	if rr.Code != code {
		return nil, fmt.Errorf("Unexpected response code %d. Expected %d", rr.Code, code)
	}

	return rr, nil
}

//helloWorldHandler is a placeholder
func helloWorldHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World.")
	}
}

//authForbidHandler returns forbidden
func authForbidHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		//Return what would be returned if access is denied.
		c.AbortWithStatus(http.StatusForbidden)
	}
}

//middlewareHandler returns a continue
func middlewareHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		//Random response code picked which returns immediately and is easy to test for
		c.AbortWithStatus(http.StatusContinue)
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

func TestAuthHandlerForbidsRequest(t *testing.T) {
	cfg := Config{
		ListenAddress: ":8888",
		Handlers:      []Handler{{Method: AnyMethod, Handler: helloWorldHandler(), Path: testEndpoint}},
		Auth:          authForbidHandler(),
	}
	_, err := checkResponseCode("GET", testEndpoint, cfg, http.StatusForbidden)
	if err != nil {
		t.Error(err)
	}
}

func TestReadinessHandler(t *testing.T) {
	cfg := Config{
		ListenAddress:  ":8888",
		Handlers:       []Handler{{Method: AnyMethod, Handler: helloWorldHandler(), Path: testEndpoint}},
		Auth:           authForbidHandler(),
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
		MiddlewareHandlers: []MiddlewareHandler{{Handler: middlewareHandler()}},
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
