// Package service provides service-related facilities.
package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/cnjack/throttle"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/puppetlabs/go-libs/internal/log"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

const (
	// AnyMethod should be passed when a handler wants to support any HTTP method.
	AnyMethod = "Any"
	// ReadinessEndpoint is the default URL for a readiness endpoint.
	ReadinessEndpoint = "/readiness"
)

// Config will hold the configuration of the service.
type Config struct {
	ListenAddress      string                   // Address in the format [host/ip]:port. Mandatory.
	LogLevel           string                   // INFO,FATAL,ERROR,WARN, DEBUG, TRACE.
	Cors               *CorsConfig              // Optional cors config.
	ReadinessCheck     bool                     // Set to true to add a readiness handler at /readiness.
	Handlers           []Handler                // Array of handlers.
	CertConfig         *ServerCertificateConfig // Optional TLS configuration.
	RateLimit          *RateLimitConfig         // Optional rate limiting config.
	MiddlewareHandlers []MiddlewareHandler      // Optional middleware handlers which will be run on every request.
	Metrics            bool                     // Optional. If true add a prometheus endpoint.
	ErrorHandler       *MiddlewareHandler       // Optional. If true a handler will be added to the end of the chain.
}

// Handler will hold all the callback handlers to be registered. N.B. gin will be used.
type Handler struct {
	Method          string                  // HTTP method or service.AnyMethod to support all limits.
	Path            string                  // The path the endpoint runs on.
	Group           string                  // Optional - specify a group (used to control which middlewares will run)
	Handler         func(c *gin.Context)    // The handler to be used.
	RateLimitConfig *HandlerRateLimitConfig // Optional rate limiting config specifically for the handler.
}

// MiddlewareHandler will hold a middleware handler and the groups on which it should be registered.
type MiddlewareHandler struct {
	Groups  []string             // Optional - what group should this middleware run on. Empty means the default route.
	Handler func(c *gin.Context) // The handler to be used.
}

// ServerCertificateConfig holds detail of the certificate config to be used.
type ServerCertificateConfig struct {
	CertificateFile string // The TLS certificate file.
	KeyFile         string // The TLS private key file.
}

// RateLimitConfig specifies the rate limiting config.
type RateLimitConfig struct {
	Groups []string // Optional - which group(s) should the rate limiting run on. Empty means the default route.
	Limit  int      // The number of requests allowed within the timeframe.
	Within int      // The timeframe(seconds) the requests are allowed in.
}

// CorsConfig specifies the CORS related config.
type CorsConfig struct {
	Groups      []string     // Optional - which group(s) should the CORS config run on. Empty means the default route.
	Enabled     bool         // Whether CORS is enabled or not.
	OverrideCfg *cors.Config // Optional. Only required if you do not want to use the default CORS configuration.
}

// HandlerRateLimitConfig holds the rate limiting config fo a sepecific handler.
type HandlerRateLimitConfig struct {
	Limit  int // The number of requests allowed within the timeframe.
	Within int // The timeframe(seconds) the requests are allowed in.
}

// Service will be the actual structure returned.
type Service struct {
	*http.Server         // Anonymous embedded struct to allow access to http server methods.
	config       *Config // The config.
}

var (
	errNoHandlersRegisteredForService = errors.New("no handlers registered for service")
	errInvalidListenAddress           = errors.New("invalid listen address")
	errRecoveredFromPanic             = errors.New("recovered from panic")
)

var routerMap = make(map[string]*gin.RouterGroup)

// Defining the readiness handler for potential use by k8s.
func readinessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	}
}

// Optional rate limiting handler.
func rateLimitHandler(limit int, within int) gin.HandlerFunc {
	return throttle.Policy(&throttle.Quota{
		Limit:  uint64(limit),
		Within: time.Duration(within) * time.Second,
	})
}

func setCorsOnRoute(group *gin.RouterGroup, overrideConfig *cors.Config) {
	if overrideConfig != nil {
		group.Use(cors.New(*overrideConfig))
	} else {
		group.Use(cors.Default())
	}
}

func setupCors(engine *gin.Engine, config *CorsConfig) {
	if config != nil {
		if config.Enabled {
			var corsGroup *gin.RouterGroup
			if len(config.Groups) == 0 {
				corsGroup = &engine.RouterGroup
				setCorsOnRoute(corsGroup, config.OverrideCfg)
			} else {
				for _, rlGroupLabel := range config.Groups {
					corsGroup = getRouterGroup(engine, rlGroupLabel)
					setCorsOnRoute(corsGroup, config.OverrideCfg)
				}
			}
		}
	}
}

func getRouterGroup(engine *gin.Engine, handlerGroup string) *gin.RouterGroup {
	if handlerGroup == "" {
		return &engine.RouterGroup
	}
	routeGroup, found := routerMap[handlerGroup]
	if found {
		return routeGroup
	}

	newGroup := engine.Group("/")
	routerMap[handlerGroup] = newGroup

	return newGroup
}

func setupRateLimiting(config *RateLimitConfig, engine *gin.Engine) {
	if config != nil {
		if len(config.Groups) == 0 {
			engine.RouterGroup.Use(rateLimitHandler(config.Limit, config.Within))
		} else {
			for _, rlGroupLabel := range config.Groups {
				rlGroup := getRouterGroup(engine, rlGroupLabel)
				rlGroup.Use(rateLimitHandler(config.Limit, config.Within))
			}
		}
	}
}

func getRateLimitHandler(config *HandlerRateLimitConfig) gin.HandlerFunc {
	return rateLimitHandler(config.Limit, config.Within)
}

func setupMiddleware(mwHandlers []MiddlewareHandler, engine *gin.Engine) {
	// Add middleware first then the handlers
	for _, handler := range mwHandlers {
		if len(handler.Groups) == 0 {
			engine.RouterGroup.Use(handler.Handler)
		} else {
			for _, handlerGroupLabel := range handler.Groups {
				handlerGroup := getRouterGroup(engine, handlerGroupLabel)
				handlerGroup.Use(handler.Handler)
			}
		}
	}
}

func setupErrorHandler(errorHandler MiddlewareHandler, engine *gin.Engine) {
	fn := func(c *gin.Context) {
		c.Next()

		errorHandler.Handler(c)
	}

	if len(errorHandler.Groups) == 0 {
		engine.RouterGroup.Use(fn)
	} else {
		for _, handlerGroupLabel := range errorHandler.Groups {
			handlerGroup := getRouterGroup(engine, handlerGroupLabel)
			handlerGroup.Use(fn)
		}
	}
}

func setupEndpoints(handlers []Handler, engine *gin.Engine) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w, error caught: %v", errRecoveredFromPanic, r)
		}
	}()

	for _, handler := range handlers {
		handlerGroup := getRouterGroup(engine, handler.Group)

		// Create a new group on the fly with the rate limiter as the first entry point and copy the chain of handlers.
		if handler.RateLimitConfig != nil {
			newHandlerGroup := engine.Group("/")

			newHandlerGroup.Handlers = append([]gin.HandlerFunc{getRateLimitHandler(handler.RateLimitConfig)},
				handlerGroup.Handlers...)

			handlerGroup = newHandlerGroup
		}

		switch method := handler.Method; method {
		case http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions:
			handlerGroup.Handle(method, handler.Path, handler.Handler)
		case AnyMethod:
			handlerGroup.Any(handler.Path, handler.Handler)
		default:
			logrus.Warnf("HTTP method %s unsupported.", method)
		}
	}

	return nil
}

// NewService will setup a new service based on the config and return this service.
func NewService(cfg *Config) (*Service, error) {
	// Router map only required in the context of this function
	if len(cfg.Handlers) == 0 {
		return nil, errNoHandlersRegisteredForService
	}

	if cfg.ListenAddress == "" {
		return nil, errInvalidListenAddress
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	logger := log.CreateLogger(cfg.LogLevel)
	router.Use(ginlogrus.Logger(logger))

	// Set CORS to the default if it's enabled and no override passed in.
	setupCors(router, cfg.Cors)

	if cfg.ReadinessCheck {
		// The readiness handler shouldn't need any middleware to run on it.
		routerGroup := router.Group("/")
		routerGroup.GET(ReadinessEndpoint, readinessHandler())
	}

	if cfg.Metrics {
		router.Handle(http.MethodGet, "metrics", gin.WrapH(promhttp.Handler()))
	}

	if cfg.ErrorHandler != nil {
		setupErrorHandler(*cfg.ErrorHandler, router)
	}

	setupRateLimiting(cfg.RateLimit, router)
	setupMiddleware(cfg.MiddlewareHandlers, router)

	err := setupEndpoints(cfg.Handlers, router)
	if err != nil {
		return nil, err
	}

	readHeaderTimeoutSeconds := 5
	server := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           router,
		ReadHeaderTimeout: time.Duration(readHeaderTimeoutSeconds) * time.Second,
	}

	return &Service{Server: server, config: cfg}, nil
}

func (s *Service) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	timeoutSeconds := 5
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	if s.Server != nil {
		if err := s.Shutdown(ctx); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

// Run will run the service in the foreground and exit when the server exits.
func (s *Service) Run() error {
	log.SetLogLevel(s.config.LogLevel)

	go func() {
		if s.config.CertConfig != nil {
			err := s.Server.ListenAndServeTLS(s.config.CertConfig.CertificateFile, s.config.CertConfig.KeyFile)
			if !errors.Is(err, http.ErrServerClosed) {
				logrus.Fatalf("Failed to start query service: %s\n", err)
			}
		} else if err := s.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("Failed to start query service: %s\n", err)
		}
	}()

	// We want a graceful exit
	return s.waitForShutdown()
}
