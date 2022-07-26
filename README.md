# go-libs 

This repository is intended as a resource for common Go code. It is hoped that the libraries provided can be used across projects to solve common problems, boost efficiency, and improve standards across the board.

Feel free to open a pull request or suggest improvements.

[![Go Report Card](https://goreportcard.com/badge/github.com/puppetlabs/go-libs)](https://goreportcard.com/report/github.com/puppetlabs/go-libs)
    
## Libraries Provided

| Library                            | Description                                                                                  |
|------------------------------------|----------------------------------------------------------------------------------------------|
| HTTP Service client library        |                                                                                              |
| HTTP Service generation            |                                                                                              |
| TLS Certificate generation library |                                                                                              |
| TLS certificate generation         |                                                                                              |
| Viper config loading               |                                                                                              |
| Concurrency                        | [Concurrency provides helpers for creating multi-threaded applications](docs/Concurrency.md) |

## Make Targets
| Target                | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
|-----------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `install-tools`       | Installs development tools, such as a compiler daemon and formatting utilities, as Go packages                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `generate-cert`       | Runs an interactive script. This script will write a new TLS certificate and private key to disk for the prompted cn and DNS name(s). The CA certificate may also be written to disk depending on the answers to the prompted for questions.                                                                                                                                                                                                                                                                                                                                                          |
| `generate-service`    | Runs an interactive script. This script will prompt the user for input on service name, directory, listening interface(optional)/port, whether HTTPS is required(certs will be auto generated to begin with), whether rate limiting, whether a readiness check is requited, whether metrics are required and whether cors is enabled. Based on the output of this a new service will be generated to the target directory with it's own Makefile, go dependencies, dockerfile and docker compose file. A hello world handler will be provided to get going. These will be ready to use out of the box |
| `all`                 | Builds the code after linting it. Various sub targets exist which are run as part of `make all`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |

## Generated Service
After running the generate-service make target the service will exist in the target directory. The targets below are some things which can be done post generation of a service:

| Target    | Description                                                                |
|-----------|----------------------------------------------------------------------------|
| `run`     | Runs the service locally.                                                  |
| `run-hot` | Runs the service locally in a hot-reloading mode using the `CompileDaemon` |
| `dev`     | Runs the service via Docker Compose                                        |
| `image`   | Builds the Docker image                                                    |

### Generated Service Code
Code will be generated into the directory specified upon running the script. A `main.go` file will exist under the `cmd` directory, and a `packages` directory will exist containing config and handlers. The code under the `pkg` directory will need edited to supplement configuration and to add any new handlers. N.B. See the config package for details on how to tag config and use nested structs.

## Linting and Formatting

We use the linter aggregator [golangci-lint](https://golangci-lint.run/).

Run `make install-tools` from the project root to install development tools.

`golangci-lint` cannot be installed as a Go package currently. It should instead be installed separately, such as using Homebrew.

Run `make format` from the project root to format all Go files in accordance with linter standards.

## To Do
* Create a workstack — there are quite a few things could go in here, such as workers
