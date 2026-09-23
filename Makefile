.PHONY: build test lint install clean

BIN_EXT := $(if $(filter windows,$(shell go env GOOS)),.exe,)

build:
	go build -trimpath -o bin/deutero-pp-cli$(BIN_EXT) ./cmd/deutero-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/deutero-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -trimpath -o bin/deutero-pp-mcp$(BIN_EXT) ./cmd/deutero-pp-mcp

install-mcp:
	go install ./cmd/deutero-pp-mcp

build-all: build build-mcp
