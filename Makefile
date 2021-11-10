# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

GO ?= latest
GOPATH := $(or $(GOPATH), $(shell go env GOPATH))
GORUN = env GOPATH=$(GOPATH) GO111MODULE=on go run

BIN = $(shell pwd)/build/bin
BUILD_PARAM?=install

OBJECTS=kcn

.PHONY: all test clean ${OBJECTS}

all: ${OBJECTS}

${OBJECTS}:
	$(GORUN) build/ci.go ${BUILD_PARAM} ./cmd/$@


test:
	$(GORUN) build/ci.go test

fmt:
	$(GORUN) build/ci.go fmt

clean:
	env GO111MODULE=on go clean -cache
	rm -fr build/_workspace/pkg/ $(BIN)/* build/_workspace/src/

# The devtools target installs tools required for 'go generate'.
# You need to put $BIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOFLAGS= GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOFLAGS= GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOFLAGS= GOBIN= go get -u github.com/fjl/gencodec
	env GOFLAGS= GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOFLAGS= GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'
