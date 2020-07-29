.DEFAULT_GOAL := build-all

# Basic Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOPATH=$(shell go env GOPATH)

# Binary name
BINARY_NAME=wakatime-cli

build-all: build-darwin build-linux build-windows

build-darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o ./build/darwin/amd64/$(BINARY_NAME) -v

build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -o ./build/linux/amd64/$(BINARY_NAME) -v

build-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) -o ./build/windows/amd64/$(BINARY_NAME).exe -v

# Install linter
.PHONY: install-linter
install-linter:
	hash golangci-lint 2>/dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.26.0

# Run static analysis tools, configuration in ./.golangci.yml file
.PHONY: lint
lint: install-linter
	golangci-lint run ./...

.PHONY: test
test:
	go test -cover -race ./...
