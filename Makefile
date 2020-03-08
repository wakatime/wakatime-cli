.DEFAULT_GOAL := build-all

# Basic Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
# Binary name
BINARY_NAME=wakatime-cli

ensure: dep
	dep ensure

build-all: build-darwin build-linux build-windows

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o ./build/darwin/amd64/$(BINARY_NAME) -v

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o ./build/linux/amd64/$(BINARY_NAME) -v

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o ./build/windows/amd64/$(BINARY_NAME).exe -v

dep:
ifeq (, $(shell which dep))
	go get github.com/golang/dep/cmd/dep
endif