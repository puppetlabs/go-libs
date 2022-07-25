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

all: check test test_generated_service build

dep:
	@go mod download

check: lint vet format sec
format: tools
	@echo "$(OK_COLOR)==> Checking code formating with 'goimports' tool$(NO_COLOR)"
	@goimports -l -w . || exit 1

generate-service:
	@echo "$(OK_COLOR)==> Generating new service.$(NO_COLOR)"
	@scripts/generate-service.sh || exit 1

generate-cert:
	@echo "$(OK_COLOR)==> Generating new TLS certificate.$(NO_COLOR)"
	@scripts/generate-cert.sh || exit 1

vet:
	@echo "$(OK_COLOR)==> Checking code correctness with 'go vet' tool$(NO_COLOR)"
	@go vet ./... || exit 1

lint: tools
	@echo "$(OK_COLOR)==> Checking code style with 'golint' tool$(NO_COLOR)"
	@golint -set_exit_status ./... || exit 1

# run gci and gofumpt to improve formatting of Go files
reformat:
	gci write ./
	gofumpt -l -w ./

PHONY+= sec
sec: $(GOPATH)/bin/gosec
	@echo "🔘 Checking for security problems ... (`date '+%H:%M:%S'`)"
	@sec=`gosec -quiet ./...`; \
	if [ "$$sec" != "" ]; \
	then echo "🔴 Problems found"; echo "$$sec"; exit 1;\
	else echo "✅ No problems found (`date '+%H:%M:%S'`)"; \
	fi

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

################################################################################
#  Tools and Deps                                                              #
#  The following targets install dependencies and tools required for the build #
################################################################################
tools: $(GO_BINARIES)/golint $(GO_BINARIES)/goimports $(GO_BINARIES)/CompileDaemon $(GOPATH)/bin/gosec
$(GO_BINARIES)/golint:
	@go get -u golang.org/x/lint/golint

$(GO_BINARIES)/goimports:
	@go get -u golang.org/x/tools/cmd/goimports

$(GO_BINARIES)/CompileDaemon:
	@go get github.com/githubnemo/CompileDaemon

$(GOPATH)/bin/gosec:
	@echo "🔘 Installing gosec ... (`date '+%H:%M:%S'`)"
	@go get -u github.com/securego/gosec/v2/cmd/gosec

init-hooks: # Set hooks path for this repo
	git config core.hooksPath .hooks
