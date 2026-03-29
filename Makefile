.PHONY: all build test lint clean run install tidy test-cover snapshot release

BINARY_NAME := symphony
BUILD_DIR := ./bin

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell git log -1 --format=%ci 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags="-s -w \
	-X 'github.com/username/symphony/cmd.Version=$(VERSION)' \
	-X 'github.com/username/symphony/cmd.Commit=$(COMMIT)' \
	-X 'github.com/username/symphony/cmd.BuildDate=$(DATE)'"

# Default: lint, test, then build
all: lint test build

build:
	@echo "Building Symphony CLI..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go
	@echo "Binary ready: $(BUILD_DIR)/$(BINARY_NAME)"

run:
	@go run $(LDFLAGS) ./main.go $(ARGS)

test:
	@go test ./... -race -v

test-cover:
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "Coverage report: coverage.html"

lint:
	@golangci-lint run ./...

clean:
	@rm -rf $(BUILD_DIR) coverage.out coverage.html dist/
	@echo "Cleaned."

snapshot:
	@goreleaser release --snapshot --clean

release:
	@goreleaser release --clean

install:
	@go install $(LDFLAGS) ./main.go

tidy:
	@go mod tidy
	@echo "go.mod and go.sum updated."
