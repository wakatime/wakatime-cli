.DEFAULT_GOAL := build-all

# globals
BINARY_NAME=wakatime-cli
COMMIT?=$(shell git rev-parse --short HEAD)
DATE?=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
REPO=github.com/wakatime/wakatime-cli
VERSION?=<local-build>

# ld flags for go build
LD_FLAGS=-s -w -X ${REPO}/pkg/version.BuildDate=${DATE} -X ${REPO}/pkg/version.Commit=${COMMIT} -X ${REPO}/pkg/version.Version=${VERSION}

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
build-all: build-all-darwin build-all-linux build-all-windows

build-all-darwin: build-darwin-amd64 build-darwin-arm64

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 make build-binary

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 make build-binary

build-all-linux: build-linux-386 build-linux-amd64 build-linux-arm build-linux-arm64

build-linux-386:
	GOOS=linux GOARCH=386 make build-binary

build-linux-amd64:
	GOOS=linux GOARCH=amd64 make build-binary

build-linux-arm:
	GOOS=linux GOARCH=arm make build-binary

build-linux-arm64:
	GOOS=linux GOARCH=arm64 make build-binary

build-all-windows: build-windows-386 build-windows-amd64

build-windows-386:
	GOOS=windows GOARCH=386 make build-binary-windows

build-windows-amd64:
	GOOS=windows GOARCH=amd64 make build-binary-windows

build-binary:
	CGO_ENABLED="0" GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -v \
		-ldflags "${LD_FLAGS} -X ${REPO}/pkg/version.OS=$(GOOS) -X ${REPO}/pkg/version.Arch=$(GOARCH)" \
		-o ./build/$(BINARY_NAME)-$(GOOS)-$(GOARCH)

build-binary-windows:
	CGO_ENABLED="0" GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -v \
		-ldflags "${LD_FLAGS} -X ${REPO}/pkg/version.OS=$(GOOS) -X ${REPO}/pkg/version.Arch=$(GOARCH)" \
		-o ./build/$(BINARY_NAME)-$(GOOS)-$(GOARCH).exe

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
	go test -cover -race ./...
