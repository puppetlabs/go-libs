
  
# go-libs 
The go-libs repo is a repo intended for the sharing of common code for use within Puppet (at present). The idea behind this is that it can be used across teams to solve common problems with the hope being that getting an uptake on this will improve efficiency and improve standards across the board.  
    
[![Go Report Card](https://goreportcard.com/badge/github.com/puppetlabs/go-libs)](https://goreportcard.com/report/github.com/puppetlabs/go-libs)
    
## Current libraries  

| Library        | Description  |
| ------------- |-------------|
| HTTP Service client library.   |          |
| HTTP Service generation       |          |
| TLS Certificate generation library. |    |
| TLS certificate generation. |    |
| Viper config loading. |    |
| Concurrency |  [Concurrecny provides helpers for creating mulit-threaded applications](docs/Concurrency.md)  |
    
## Make targets
| Target |Description  |
|--|--|
|generate-cert  |**generate-cert** will run an interactive script. This script will write a new TLS certificate and private key to disk for the prompted cn and DNS name(s). The CA certificate may also be written to disk depending on the answers to the prompted for questions.   |
|generate-service  |**generate-service** will run an interactive script. This script will prompt the user for input on service name, directory, listening interface(optional)/port, whether HTTPS is required(certs will be auto generated to begin with), whether rate limiting, whether a readiness check is requited, whether metrics are required and whether cors is enabled. Based on the output of this a new service will be generated to the target directory with it's own Makefile, go dependencies, dockerfile and docker compose file. A hello world handler will be provided to get going. These will be ready to use out of the box |
|all  |all will build the code after linting it, security linting it and vetting it. Various sub targets exist which are run as part of make all. |

#### Generated service
After running the generate-service make target the service will exist in the target directory. The targets below are some things which can be done post generation of a service:
|  Target|Description  |
|--|--|
|  run|  Runs the service locally.|
| run-hot |Runs the service locally using the CompileDaemon. This will mean that hot reloading will happen.  |
| dev |Runs the service in docker-compose  |
| image |Builds the docker image.  |

##### Generated Service Code
Code will be generated into the directory specified upon running the script. A main.go file will exist under the cmd directory and a packages directory will exist containing config and handlers. The code under the pkg directory will need edited to supplement configuration and to add any new handlers. N.B. See the config package for details on how to tag config and use nested structs.
    
**TODO:**  
- Create a workstack - there are quite a few things could go in here like workers