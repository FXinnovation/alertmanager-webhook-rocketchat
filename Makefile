
SRC_DIR=github.com/fxinnovation/alertmanager-webhook-rocketchat
BUILD_VERSION=$(shell cat VERSION.txt)
APPL=alertmanager-webhook-rocketchat

######## commom

PKGGOFILES=$(shell go list ./... | grep -v /vendor/)

GIT_COMMIT?=$(shell git rev-parse --short HEAD)
GIT_DIRTY?=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
GIT_DESCRIBE?=$(shell git describe --tags --always)
BUILD_TIME?=$(shell date +"%Y-%m-%dT%H:%M:%S")

LDFLAGS=-ldflags "\
          -X $(SRC_DIR)/information.Version=$(BUILD_VERSION) \
          -X $(SRC_DIR)/information.BuildTime=$(BUILD_TIME) \
          -X $(SRC_DIR)/information.GitCommit=$(GIT_COMMIT) \
          -X $(SRC_DIR)/information.GitDirty=$(GIT_DIRTY) \
          -X $(SRC_DIR)/information.GitDescribe=$(GIT_DESCRIBE)"


PWD=$(shell pwd)

.PHONY: all help clean test test-cover test-coverage dependencies build fmt vet lint tools

all: fmt build test

help:
	@grep -hE '^[a-zA-Z_-]+.*?:.*?## .*$$' ${MAKEFILE_LIST} | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[0;49;95m%-30s\033[0m %s\n", $$1, $$2}'

## If you have go on your wonderful laptop
clean:
	@rm -rf ./target || true
	@mkdir ./target || true

test: fmt vet ## go test
	go test -cpu=2 -p=2 -v --short $(LDFLAGS) $(PKGGOFILES)

test-cover: fmt vet ## go test with coverage
	go test  $(PKGGOFILES) -cover -race -v $(LDFLAGS)

test-coverage: clean fmt vet ## for jenkins
	gocov test $(PKGGOFILES) --short -cpu=2 -p=2 -v $(LDFLAGS) | gocov-xml > ./coverage-test.xml

dependencies: ## download the dependencies
	rm -rf Gopkg.lock vendor/
	dep ensure

build: clean fmt vet
	go build $(LDFLAGS)

fmt: ## go fmt on packages
	go fmt $(PKGGOFILES)

vet: ## go vet on packages
	go vet $(PKGGOFILES)

lint: ## go vet on packages
	golint -set_exit_status=true $(PKGGOFILES)

tools: ## install tools to develop
	go get -u github.com/golang/dep/cmd/dep
	go get -u golang.org/x/lint/golint
	go get github.com/axw/gocov/...
	go get github.com/AlekSi/gocov-xml
