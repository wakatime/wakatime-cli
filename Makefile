.DEFAULT_GOAL := build-all

# globals
BINARY_NAME=wakatime-cli
COMMIT?=$(shell git rev-parse --short HEAD)
DATE?=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
REPO=github.com/wakatime/wakatime-cli

# ld flags for go build
LD_FLAGS=-s -w -X ${REPO}/pkg/version.BuildDate=${DATE} -X ${REPO}/pkg/version.Commit=${COMMIT} -X ${REPO}/pkg/version.Version=<local-build>

# basic Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# linting
define get_latest_lint_release
	curl -s "https://api.github.com/repos/golangci/golangci-lint/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
endef
LATEST_LINT_VERSION=$(shell $(call get_latest_lint_release))
INSTALLED_LINT_VERSION=$(shell golangci-lint --version 2>/dev/null | awk '{print "v"$$4}')

# get GOPATH according to OS
ifeq ($(OS),Windows_NT) # is Windows_NT on XP, 2000, 7, Vista, 10...
    GOPATH=$(go env GOPATH)
else
    GOPATH=$(shell go env GOPATH)
endif

# targets
build-all: build-darwin build-linux build-windows

build-darwin: generate
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) -v \
		-ldflags "${LD_FLAGS} -X ${REPO}/pkg/version.OS=darwin -X ${REPO}/pkg/version.Arch=amd64" \
		-o ./build/darwin/amd64/$(BINARY_NAME)

build-linux: generate
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -v \
		-ldflags "${LD_FLAGS} -X ${REPO}/pkg/version.OS=linux -X ${REPO}/pkg/version.Arch=amd64" \
		-o ./build/linux/amd64/$(BINARY_NAME)

build-windows: generate
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) -v \
		-ldflags "${LD_FLAGS} -X ${REPO}/pkg/version.OS=windows -X ${REPO}/pkg/version.Arch=amd64" \
		-o ./build/windows/amd64/$(BINARY_NAME).exe

# generate plugin language mapping code
.PHONY: generate
generate:
	go run ./cmd/generate/main.go

# install linter
.PHONY: install-linter
install-linter:
ifneq "$(INSTALLED_LINT_VERSION)" "$(LATEST_LINT_VERSION)"
	@echo "new golangci-lint version found:" $(LATEST_LINT_VERSION)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin latest
endif

# run static analysis tools, configuration in ./.golangci.yml file
.PHONY: lint
lint: install-linter
	golangci-lint run ./...

.PHONY: test
test: generate
	go test -race -covermode=atomic -coverprofile=coverage.out ./...
