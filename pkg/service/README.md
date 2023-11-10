
# service  
The intention behind the service package is to allow a client to  get a service up and running by passing in a configuration struct.  
    
## Supported features
- readiness handler (can be required by k8s) 
- CORS - default configuration if enabled or supplied override configuration  
- Logging.  
- The default prometheus metrics endpoint.  
- Listening on HTTP or HTTPS.  
- Rate limiting.  
- Adding new handlers.  
- Adding new middleware.  
- Adding an auth handler.  
  
### API
```
//NewService will setup a new service based on the config and return this service.  
func NewService(cfg *Config) (*Service, error)

//Run will run the service in the foreground and exit when the HTTP server exits  
func (s *Service) Run() error
```

### Types
```
//Config will hold the configuration of the service.  
type Config struct {  
  ListenAddress      string                   //Address in the format [host/ip]:port. Mandatory  
  LogLevel           string                   //INFO,FATAL,ERROR,WARN, DEBUG, TRACE  
  Cors               *CorsConfig              //Optional cors config  
  ReadinessCheck     bool                     //Set to true to add a readiness handler at /readiness.  
  Handlers           []Handler                //Array of handlers. N.B. At least one handler is required.  
  CertConfig         *ServerCertificateConfig //Optional TLS configuration  
  RateLimit          *RateLimitConfig         //Optional rate limiting config  
  MiddlewareHandlers []MiddlewareHandler      //Optional middleware handlers which will be run on every request  
  Metrics            bool                     //Optional. If true a prometheus metrics endpoint will be exposed at /metrics/  
  ErrorHandler       *MiddlewareHandler       //Optional. If true a handler will be added to the end of the chain.
}  
  
// Handler will hold all the callback handlers to be registered. N.B. gin will be used.
type Handler struct {
	Method          string                  // HTTP method or service.AnyMethod to support all limits.
	Path            string                  // The path the endpoint runs on.
	Group           string                  // Optional - specify a group (used to control which middlewares will run)
	Handler         func(c *gin.Context)    // The handler to be used.
	RateLimitConfig *HandlerRateLimitConfig // Optional rate limiting config specifically for the handler.
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

// HandlerRateLimitConfig holds the rate limiting config fo a sepecific handler.
type HandlerRateLimitConfig struct {
	Limit  int // The number of requests allowed within the timeframe.
	Within int // The timeframe(seconds) the requests are allowed in.
}
  
//Service will be the actual structure returned.  
type Service struct {  
  *http.Server //Anonymous embedded struct to allow access to http server methods.  
  config *Config //The config.  
}
```

#### Notes
- The cors config and the handlers are based on the gin framework : https://github.com/gin-gonic/gin.  
- Rate limiting is done by the library github.com/cnjack/throttle.
- Rate limiting can be added to a handler or on a per group basis. 
- The group principle is based on Gin routergroups. The idea behind it is that not all middleware needs to run on 
all requests so the middleware in a group will only run against an endpoint in that group. 
This is applied to cors, rate limiting and any middleware in general.  
- See internal/examples/service/main.go for an example of how to use the service package to generate a service.  
    
    
**TODO:** - Consider adding GRPC.    
- Consider adding GraphQL.  
- Flesh out more with logging.  
- Look at potentially not using gin (it does the work for you so not a high priority).