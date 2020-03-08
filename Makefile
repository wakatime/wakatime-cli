.DEFAULT_GOAL := build

ensure: dep
	dep ensure

build:
	go build -o ./build/${GOOS}/${GOARCH}/wakatime-cli .

dep:
ifeq (, $(shell which dep))
	go get github.com/golang/dep/cmd/dep
endif