COMMIT=$(shell git describe --always)
NOW=$(shell date +'%Y-%m-%d_%T')
TEST_SERVICE=test-generated-service

all: lint test build test-generated-service

# install the necessary development tools
PHONY+= install-tools
install-tools:
	@go install github.com/4meepo/tagalign/cmd/tagalign@latest
	@go install github.com/daixiang0/gci@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
	@go install github.com/ypresto/CompileDaemon@bump-sys-fork # using fork until bug resolved: https://github.com/githubnemo/CompileDaemon/pull/76
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install mvdan.cc/gofumpt@latest

# run formatting utilities
PHONY+= format
format:
	@goimports -l ./
	@gci write ./
	@gofumpt -l -w ./
	@tagalign -fix -sort ./...

# run linters
PHONY+= lint
lint:
	@golangci-lint run

# run unit tests
PHONY+= test
test:
	@echo "$(OK_COLOR)==> Running tests$(NO_COLOR)"
	@go test -v -race -cover ./... || exit 1

PHONY+= generate-cert
generate-cert:
	@echo "$(OK_COLOR)==> Generating new TLS certificate.$(NO_COLOR)"
	@scripts/generate-cert.sh || exit 1

PHONY+= generate-service
generate-service:
	@echo "$(OK_COLOR)==> Generating new service.$(NO_COLOR)"
	@scripts/generate-service.sh || exit 1

# ensures code changes have not broken the templates – no actual tests are run
PHONY+= test-generated-service
test-generated-service:
	@echo "$(OK_COLOR)==>Testing generated service.$(NO_COLOR)"
	@mkdir -p $(PWD)/$(TEST_SERVICE)
	@go run -ldflags "-X main.name=$(TEST_SERVICE) -X main.serviceDir=$(PWD)/$(TEST_SERVICE) -X main.listenAddress=:8888 -X main.listenPort=8888" cmd/service-generator/main.go
	@go mod tidy
	@go get github.com/gin-gonic/gin
	@cd $(PWD)/$(TEST_SERVICE);go test ./...
	@rm -rf $(PWD)/$(TEST_SERVICE)

PHONY+= build
build:
	@echo "$(OK_COLOR)==> Building$(NO_COLOR)"
	@CGO_ENABLED=0 go build -ldflags "-X main.sha1ver=${COMMIT} -X main.buildTime=${NOW}" -a -installsuffix cgo ./... || exit 1

.PHONY: $(PHONY)
