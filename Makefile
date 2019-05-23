GO    := GO15VENDOREXPERIMENT=1 go
PROMU := $(GOPATH)/bin/promu
pkgs   = $(shell $(GO) list ./... | grep -v /vendor/)

PREFIX                  ?= $(shell pwd)
BIN_DIR                 ?= $(shell pwd)
DOCKER_REPO             ?= fxinnovation
DOCKER_IMAGE_NAME       ?= alertmanager-webhook-rocketchat
DOCKER_IMAGE_TAG        ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))

all: vet format build test

test: build ## running test after build
	@echo ">> running tests"
	@$(GO) test -cpu=2 -p=2 -v -short $(pkgs)

test-cover: format vet ## go test with coverage
	go test  $(pkgs) -cover -race -v

style: ## check code style
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

format: ## Format code
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet: ## vet code
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

dependencies: tools ## download the dependencies
	rm -rf Gopkg.lock vendor/
	dep ensure

tools: ## install tools to develop
	go get -u github.com/golang/dep/cmd/dep
	go get github.com/axw/gocov/...

build: promu ## build code with promu
	@echo ">> building binaries"
	@$(PROMU) build --prefix $(PREFIX)

tarball: promu ## creates a release tarball
	@echo ">> building release tarball"
	@$(PROMU) tarball --prefix $(PREFIX) $(BIN_DIR)

docker: ## creates docker image
	@echo ">> building docker image"
	@docker build -t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

dockerlint: ## lints dockerfile
	@echo ">> linting Dockerfile"
	@docker run --rm -i hadolint/hadolint < Dockerfile

promu: ## gets promu for building
	@GOOS=$(shell uname -s | tr A-Z a-z) \
		GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
		$(GO) get -u github.com/prometheus/promu

lint: golint ## lint code
	@echo ">> linting code"
	@! golint $(pkgs) | grep '^'

golint: ## gets golint for building
	@go get -u golang.org/x/lint/golint

help:
	@grep -hE '^[a-zA-Z_-]+.*?:.*?## .*$$' ${MAKEFILE_LIST} | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[0;49;95m%-30s\033[0m %s\n", $$1, $$2}'


.PHONY: all style format dependencies build test vet tarball promu


VERSION?=$(shell cat VERSION.txt)
TAGEXISTS=$(shell git tag --list | egrep -q "^$(VERSION)$$" && echo 1 || echo 0)

release-tag: ## create a release tag
	@if [ $(TAGEXISTS) == 0 ]; then \
		echo ">>Creating tag $(VERSION)";\
		git tag $(VERSION);\
        git push -u origin $(VERSION);\
	fi

release-version: release-tag release-docker ## tag the version with VERSION.txt

release: ## creates a crossbuild in .build folder for either all the supported platforms or the one passed as OSARCH var
	@if [ -n "$(OSARCH)" ]; then \
		export PLATFORMS="-p $(OSARCH)";\
	fi
	echo $(PLATFORMS)
	@$(PROMU) crossbuild $(PLATFORMS)

release-docker: ## pushes the VERSION.txt tag to docker hub
	@docker build -t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)" .
	@docker push $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)

