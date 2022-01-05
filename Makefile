#!/usr/bin/make -f

VERSION := $(shell echo $(shell git describe --always --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
BuildTime :=$(shell date '+%Y-%m-%dT%H:%M:%SZ%z')
ldflags = '-X github.com/functionx/fx-tron-bridge.Version=$(VERSION) \
           -X github.com/functionx/fx-tron-bridge.Commit=$(COMMIT) \
           -X github.com/functionx/fx-tron-bridge.BuildTime=$(BuildTime) \
           -w -s'

###############################################################################
###                              Documentation                              ###
###############################################################################

.PHONY: build install go.sum

BUILDDIR ?= $(CURDIR)/build

build: go.mod
	@go build -mod=readonly -ldflags $(ldflags) -v -o $(BUILDDIR)/bin/fxtronbridge ./cmd

build-linux:
	@CGO_ENABLED=0 TARGET_CC=clang LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 make build

install:
	@$(MAKE) build
	@mv $(BUILDDIR)/bin/fxtronbridge $(GOPATH)/bin/fxtronbridge

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify
	@go mod tidy
	@echo "--> Download go modules to local cache"
	@go mod download

docker: build-linux
	@docker build -t functionx/fx-tron-bridge:latest .

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	@echo "--> Running linter"
	@golangci-lint run
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.*' | xargs gofmt -d -s

format:
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.*' | xargs gofmt -w -s
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.*' | xargs misspell -w
