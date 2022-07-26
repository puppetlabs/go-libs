REGISTRY = 'gcr.io'
PROJECT = 'estate-reporting'
COMPONENT = 'go-service'
TAG := $(or ${TAG},${TAG},dev-local)
GO_BINARIES := $(HOME)/go
ifdef GOPATH
	GO_BINARIES = $(GOPATH)/bin
endif
COMMIT=$(shell git describe --always)
now=$(shell date +'%Y-%m-%d_%T')
TEST_SERVICE=test-generated-service

all: test test_generated_service build

dep:
	@go mod download

generate-service:
	@echo "$(OK_COLOR)==> Generating new service.$(NO_COLOR)"
	@scripts/generate-service.sh || exit 1

generate-cert:
	@echo "$(OK_COLOR)==> Generating new TLS certificate.$(NO_COLOR)"
	@scripts/generate-cert.sh || exit 1

test:
	@echo "$(OK_COLOR)==> Running tests$(NO_COLOR)"
	@go test -v -race -cover ./... || exit 1

#This target is to make sure code changes have not broken the templates - no actual tests will be run.
test_generated_service:
	@echo "$(OK_COLOR)==>Testing genenerated service.$(NO_COLOR)"
	@mkdir -p $(PWD)/$(TEST_SERVICE)
	@go run -ldflags "-X main.name=$(TEST_SERVICE) -X main.serviceDir=$(PWD)/$(TEST_SERVICE) -X main.listenAddress=:8888 -X main.listenPort=8888" cmd/service-generator/main.go
	@cd $(PWD)/$(TEST_SERVICE);go test ./...
	@rm -rf $(PWD)/$(TEST_SERVICE)

build:
	@echo "$(OK_COLOR)==> Building$(NO_COLOR)"
	@CGO_ENABLED=0 go build -ldflags "-X main.sha1ver=${COMMIT} -X main.buildTime=${now}" -a -installsuffix cgo ./... || exit 1

# installs development tools, such as a compiler daemon and formatting utilities, as Go packages
install-tools:
	@go install github.com/githubnemo/CompileDaemon@latest
	@go install github.com/daixiang0/gci@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install mvdan.cc/gofumpt@v0.3.0	# v0.3.1 fails to install on Go 1.16

# run linters
lint:
	@golangci-lint run

# run formatting utilities to improve formatting of Go files
format:
	goimports -l ./
	gci write ./
	gofumpt -l -w ./

init-hooks: # Set hooks path for this repo
	git config core.hooksPath .hooks
