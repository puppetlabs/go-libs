
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
  Auth               func(c *gin.Context)     //An optional Auth handler.  
  ReadinessCheck     bool                     //Set to true to add a readiness handler at /readiness.  
  Handlers           []Handler                //Array of handlers. N.B. At least one handler is required.  
  CertConfig         *ServerCertificateConfig //Optional TLS configuration  
  RateLimit          *RateLimitConfig         //Optional rate limiting config  
  MiddlewareHandlers []MiddlewareHandler      //Optional middleware handlers which will be run on every request  
  Metrics            bool                     //Optional. If true a prometheus metrics endpoint will be exposed at /metrics/  
}  
  
//Handler will hold all the callback handlers to be registered. N.B. gin will be used.  
type Handler struct {  
  Method            string               //HTTP method or service.AnyMethod to support all limits.  
  Path              string               //The path the endpoint runs on.  
  OverrideRateLimit bool                 //Optional - set to true if rate limiting is on and this handler will not use it.  
  Handler           func(c *gin.Context) //The handler to be used.  
}  
  
//MiddlewareHandler will hold all the middleware and whether  
type MiddlewareHandler struct {  
  OverrideRateLimit bool                 //Optional - set to true if rate limiting is on and this handler will not use it.  
  Handler           func(c *gin.Context) //The handler to be used.  
}  
  
//ServerCertificateConfig holds detail of the certificate config to be used  
type ServerCertificateConfig struct {  
  CertificateFile string //The TLS certificate file.  
  KeyFile         string //The TLS private key file.  
}  
  
//RateLimitConfig specifies the rate limiting config  
type RateLimitConfig struct {  
  Limit  int //The number of requests allowed within the timeframe.  
  Within int //The timeframe(seconds) the requests are allowed in.  
}  
  
//CorsConfig specifies the CORS related config  
type CorsConfig struct {  
  Enabled     bool //Whether CORS is enabled or not.  
  OverrideCfg *cors.Config //Optional. This is only required if you do not want to use the default CORS configuration.  
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
- See internal/examples/service/main.go for an example of how to use the service package to generate a service.  
    
    
**TODO:** - Consider adding GRPC.    
- Consider adding GraphQL.  
- Flesh out more with logging.  
- Look at potentially not using gin (it does the work for you so not a high priority).