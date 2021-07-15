package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cnjack/throttle"

	"github.com/puppetlabs/go-libs/internal/log"
	ginlogrus "github.com/toorop/gin-logrus"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	//AnyMethod should be passed when a handler wants to support any HTTP method.
	AnyMethod = "Any"
	//ReadinessEndpoint is the default URL for a readiness endpoint
	ReadinessEndpoint = "/readiness"
)

//Config will hold the configuration of the service.
type Config struct {
	ListenAddress      string                   //Address in the format [host/ip]:port. Mandatory
	LogLevel           string                   //INFO,FATAL,ERROR,WARN, DEBUG, TRACE
	Cors               *CorsConfig              //Optional cors config
	ReadinessCheck     bool                     //Set to true to add a readiness handler at /readiness.
	Handlers           []Handler                //Array of handlers
	CertConfig         *ServerCertificateConfig //Optional TLS configuration
	RateLimit          *RateLimitConfig         //Optional rate limiting config
	MiddlewareHandlers []MiddlewareHandler      //Optional middleware handlers which will be run on every request
	Metrics            bool                     //Optional. If true a prometheus metrics endpoint will be exposed at /metrics/
}

//Handler will hold all the callback handlers to be registered. N.B. gin will be used.
type Handler struct {
	Method  string               //HTTP method or service.AnyMethod to support all limits.
	Path    string               //The path the endpoint runs on.
	Group   string               //Optional - specify a group if this is to have it's own group. N.B. The point of the group is to allow middleware to run on some requests and not others(based on the group).
	Handler func(c *gin.Context) //The handler to be used.
}

//MiddlewareHandler will hold all the middleware and whether
type MiddlewareHandler struct {
	Groups  []string             //Optional - what group should this middleware run on. Empty means the default route.
	Handler func(c *gin.Context) //The handler to be used.
}

//ServerCertificateConfig holds detail of the certificate config to be used
type ServerCertificateConfig struct {
	CertificateFile string //The TLS certificate file.
	KeyFile         string //The TLS private key file.
}

//RateLimitConfig specifies the rate limiting config
type RateLimitConfig struct {
	Groups []string //Optional - which group(s) should the rate limiting run on. Empty means the default route.
	Limit  int      //The number of requests allowed within the timeframe.
	Within int      //The timeframe(seconds) the requests are allowed in.
}

//CorsConfig specifies the CORS related config
type CorsConfig struct {
	Groups      []string     //Optional - which group(s) should the CORS config run on. Empty means the default route.
	Enabled     bool         //Whether CORS is enabled or not.
	OverrideCfg *cors.Config //Optional. This is only required if you do not want to use the default CORS configuration.
}

//Service will be the actual structure returned.
type Service struct {
	*http.Server         //Anonymous embedded struct to allow access to http server methods.
	config       *Config //The config.
}

var routerMap = make(map[string]*gin.RouterGroup)

// Defining the readiness handler for potential use by k8s
func readinessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	}
}

//Optional rate limiting handler
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

func setupMiddleware(mwHandlers []MiddlewareHandler, engine *gin.Engine) {
	//Add middleware first then the handlers
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

func setupEndpoints(handlers []Handler, engine *gin.Engine) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error caught: %s", r)
		}
	}()

	for _, handler := range handlers {
		handlerGroup := getRouterGroup(engine, handler.Group)
		switch method := handler.Method; method {
		case http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete:
			handlerGroup.Handle(method, handler.Path, handler.Handler)
		case AnyMethod:
			handlerGroup.Any(handler.Path, handler.Handler)
		default:
			logrus.Warnf("HTTP method %s unsupported.", method)
		}
	}
	return nil
}

//NewService will setup a new service based on the config and return this service.
func NewService(cfg *Config) (*Service, error) {
	//Router map only required in the context of this function
	if len(cfg.Handlers) == 0 {
		return nil, fmt.Errorf("no handlers registered for service")
	}

	if cfg.ListenAddress == "" {
		return nil, fmt.Errorf("listen address must be valid")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	logger := log.CreateLogger(cfg.LogLevel)
	router.Use(ginlogrus.Logger(logger))

	//Set CORS to the default if it's enabled and no override passed in.
	setupCors(router, cfg.Cors)

	if cfg.ReadinessCheck {
		//The readiness handler shouldn't need any middleware to run on it.
		routerGroup := router.Group("/")
		routerGroup.GET(ReadinessEndpoint, readinessHandler())
	}

	if cfg.Metrics {
		router.Handle(http.MethodGet, "metrics", gin.WrapH(promhttp.Handler()))
	}

	setupRateLimiting(cfg.RateLimit, router)
	setupMiddleware(cfg.MiddlewareHandlers, router)
	err := setupEndpoints(cfg.Handlers, router)

	if err != nil {
		return nil, err
	}

	server := &http.Server{
		Addr:    cfg.ListenAddress,
		Handler: router,
	}

	return &Service{Server: server, config: cfg}, nil
}

func (s *Service) waitForShutdown() error {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if s.Server != nil {
		if err := s.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}

//Run will run the service in the foreground and exit when the server exits
func (s *Service) Run() error {
	log.SetLogLevel(s.config.LogLevel)

	go func() {
		if s.config.CertConfig != nil {
			if err := s.Server.ListenAndServeTLS(s.config.CertConfig.CertificateFile, s.config.CertConfig.KeyFile); err != http.ErrServerClosed {
				logrus.Fatalf("Failed to start query service: %s\n", err)
			}
		} else {
			if err := s.Server.ListenAndServe(); err != http.ErrServerClosed {
				logrus.Fatalf("Failed to start query service: %s\n", err)
			}
		}
	}()

	//We want a graceful exit
	if err := s.waitForShutdown(); err != nil {
		return err
	}

	return nil
}
