package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/puppetlabs/go-libs/pkg/service"
)

const errorExitCode = 1

/*The service example will show you how to create a service using the code within pkg/service.
This will illustrate gin handlers, non gin handlers and how to use them, the use of an AUTH
handler, handlers with rate limiting, handler without rate limiting and default cors configuration.
To run this perform a go run internal/examples/service/main.go from the top level directory.
*/

// BasicTest is a placeholder
func BasicTest() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Returning from standard HTTP handler ")
}

func middlewareHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		fmt.Println("In middleware")
	}
}

func main() {
	handlers := []service.Handler{
		{Method: http.MethodGet, Path: "/test", Handler: BasicTest()},
		{Method: http.MethodGet, Path: "/handler", Handler: gin.WrapF(handler)},
		{Method: http.MethodGet, Path: "/testRateLimit", Handler: BasicTest(), Group: "ratelimited"},
	}

	mwHandlers := []service.MiddlewareHandler{{Handler: middlewareHandler()}}

	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(errorExitCode)
	}
	certConfig := &service.ServerCertificateConfig{
		CertificateFile: fmt.Sprintf("%s/server.crt", wd),
		KeyFile:         fmt.Sprintf("%s/server.key", wd),
	}

	rateLimitConfig := &service.RateLimitConfig{Groups: []string{"ratelimited"}, Limit: 1, Within: 1}

	cfg := &service.Config{
		ListenAddress:      ":8888",
		Cors:               &service.CorsConfig{Enabled: true},
		Handlers:           handlers,
		CertConfig:         certConfig,
		LogLevel:           "WARN",
		ReadinessCheck:     true,
		RateLimit:          rateLimitConfig,
		MiddlewareHandlers: mwHandlers,
		Metrics:            true,
	}

	service, err := service.NewService(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(errorExitCode)
	}
	err = service.Run()
	if err != nil {
		os.Exit(errorExitCode)
	}
}
