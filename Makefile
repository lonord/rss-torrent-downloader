APP_NAME := rss-torrent-dl
APP_VERSION := 2.5
BUILD_TIME := $(shell date "+%F %T %Z")
PWD := $(shell pwd)
OUTDIR ?= $(PWD)/build
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOEXE := $(shell go env GOEXE)
GO_BUILD := go build -ldflags "-s -w" -ldflags "-X 'main.appName=$(APP_NAME)' -X 'main.appVersion=$(APP_VERSION)' -X 'main.buildTime=$(BUILD_TIME)'"

REGISTRY = lonord
DOCKER_BUILD = docker buildx build --platform=linux/amd64,linux/arm64

.PHONY: build docker clean rebuild linux/amd64 linux/arm64 linux/arm windows/amd64 darwin/amd64 darwin/arm64 linux windows darwin buildall

build:
	mkdir -p $(OUTDIR)/ && $(GO_BUILD) -o $(OUTDIR)/$(APP_NAME)_$(GOOS)_$(GOARCH)$(GOEXE) .

docker:
	$(DOCKER_BUILD) -t $(REGISTRY)/$(APP_NAME):$(APP_VERSION) -t $(REGISTRY)/$(APP_NAME) . --push

rebuild: clean build

linux/amd64: export GOOS=linux
linux/amd64: export GOARCH=amd64
linux/amd64: export GOEXE=
linux/amd64: build

linux/arm64: export GOOS=linux
linux/arm64: export GOARCH=arm64
linux/arm64: export GOEXE=
linux/arm64: build

linux/arm: export GOOS=linux
linux/arm: export GOARCH=arm
linux/arm: export GOEXE=
linux/arm: build

windows/amd64: export GOOS=windows
windows/amd64: export GOARCH=amd64
windows/amd64: export GOEXE=.exe
windows/amd64: build

darwin/amd64: export GOOS=darwin
darwin/amd64: export GOARCH=amd64
darwin/amd64: export GOEXE=
darwin/amd64: build

darwin/arm64: export GOOS=darwin
darwin/arm64: export GOARCH=arm64
darwin/arm64: export GOEXE=
darwin/arm64: build

linux:
	make linux/amd64
	make linux/arm64
	make linux/arm

windows:
	make windows/amd64

darwin:
	make darwin/amd64
	make darwin/arm64

buildall: linux windows darwin

rebuildall: clean buildall

clean:
	rm -rf $(OUTDIR)