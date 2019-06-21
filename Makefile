.DEFAULT_GOAL := help

#### Constant variables
# Set a default VERSION only if it is not already set
VERSION ?= vlatest
COMMAND := appsody-controller
BUILD_PATH := ./build
PACKAGE_PATH := ./package
GO_PATH := $(shell go env GOPATH)
GOLANGCI_LINT_BINARY := $(GO_PATH)/bin/golangci-lint
GOLANGCI_LINT_VERSION := v1.16.0
.PHONY: all
all: lint test package ## Run lint, test, build, and package

.PHONY: test
test: ## Run the automated tests
		go test -v -count=1 ./test/*
  
.PHONY: lint
lint: $(GOLANGCI_LINT_BINARY) ## Run the static code analyzers
# Configure the linter here. Helpful commands include `golangci-lint linters` and `golangci-lint run -h`
# Set exclude-use-default to true if this becomes to noisy.
	golangci-lint run -v --disable-all \
		--enable deadcode \
		--enable errcheck \
		--enable gosimple \
		--enable govet \
		--enable ineffassign \
		--enable staticcheck \
		--enable structcheck \
		--enable typecheck \
		--enable unused \
		--enable varcheck \
		--enable gofmt \
		--enable golint \
		--enable gofmt \
		--exclude-use-default=true \
		./...

# not PHONY, installs golangci-lint if it doesn't exist
$(GOLANGCI_LINT_BINARY):
	# see https://github.com/golangci/golangci-lint
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GO_PATH)/bin $(GOLANGCI_LINT_VERSION)

.PHONY: clean
clean: ## Removes existing build artifacts in order to get a fresh build
	rm -rf $(BUILD_PATH)
	rm -rf $(PACKAGE_PATH)
	rm -f $(GOLANGCI_LINT_BINARY)
	go clean

.PHONY: build
build: ## Build binary for linux stores it in the build/ dir
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_PATH)/$(COMMAND) -ldflags "-X main.VERSION=$(VERSION)"

.PHONY: package
package: build ## Build the linux binary and stores it in package/ dir
	mkdir -p $(PACKAGE_PATH)
	cp -p $(BUILD_PATH)/$(COMMAND) $(PACKAGE_PATH)/

# Auto documented help from http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Prints this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
