.PHONY: all deps binary plugins

DEP_VERSION=0.5.0
OS := $(shell uname | tr '[:upper:]' '[:lower:]')

all: deps binary plugins

prepare:
	@echo "Installing dep..."
	@curl -L -s https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-${OS}-amd64 -o ${GOPATH}/bin/dep
	@chmod a+x ${GOPATH}/bin/dep

deps:
	@echo "Setting up the vendors folder..."
	@dep ensure -v
	@echo ""
	@echo "Resolved dependencies:"
	@dep status
	@echo ""

binary:
	go build ./rs-server/redis_spy.go

plugins:
	go build -buildmode=plugin -o ./plugins/cache.so ./plugins/cache/cache.go
	go build -buildmode=plugin -o ./plugins/primary.so ./plugins/primary/primary.go
