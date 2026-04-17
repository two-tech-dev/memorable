.PHONY: build run test test-integration lint fmt vet tidy clean help

BINARY_NAME = memorable
BUILD_DIR   = bin
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE       ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS     = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

## build: Compile the binary
build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/memorable

## run: Build and run the server
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

## test: Run unit tests with race detection
test:
	go test -race -count=1 ./...

## test-integration: Run integration tests (requires PostgreSQL)
test-integration:
	go test -race -tags=integration -count=1 ./...

## test-cover: Run tests with coverage report
test-cover:
	go test -race -coverprofile=coverage.txt -covermode=atomic $$(go list ./... | grep -v '/mcp$$' | grep -v '/store$$' | grep -v '/cmd/')
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linters
lint: vet
	golangci-lint run

## vet: Run go vet
vet:
	go vet ./...

## fmt: Format all Go files
fmt:
	gofmt -s -w .

## tidy: Tidy and verify dependencies
tidy:
	go mod tidy
	go mod verify

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR) coverage.txt coverage.html

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'
