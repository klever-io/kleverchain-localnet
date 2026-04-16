BINARY_NAME := localnet
CMD_DIR     := ./cmd/localnet
BIN_DIR     := bin
DIST_DIR    := dist
COVER_FILE  := coverage.out

VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE  ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X github.com/klever-io/kleverchain-localnet/internal/version.Version=$(VERSION) \
	-X github.com/klever-io/kleverchain-localnet/internal/version.Commit=$(COMMIT) \
	-X github.com/klever-io/kleverchain-localnet/internal/version.Date=$(BUILD_DATE)

GO       ?= go
GOFLAGS  ?=

.PHONY: all build test test-race lint fmt cover install clean tidy help

all: lint test build

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)

install:
	$(GO) install $(GOFLAGS) -trimpath -ldflags "$(LDFLAGS)" $(CMD_DIR)

test:
	$(GO) test $(GOFLAGS) ./...

test-race:
	$(GO) test $(GOFLAGS) -race ./...

cover:
	$(GO) test $(GOFLAGS) -race -coverprofile=$(COVER_FILE) -covermode=atomic ./...
	$(GO) tool cover -func=$(COVER_FILE)

lint:
	golangci-lint run ./...

fmt:
	$(GO) fmt ./...
	@command -v goimports >/dev/null 2>&1 && goimports -w -local github.com/klever-io/kleverchain-localnet . || true

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR) $(COVER_FILE)

help:
	@echo "Targets:"
	@echo "  build      Build the localnet binary into $(BIN_DIR)/"
	@echo "  install    Install the localnet binary into GOBIN"
	@echo "  test       Run unit tests"
	@echo "  test-race  Run unit tests with the race detector"
	@echo "  cover      Run tests and report coverage"
	@echo "  lint       Run golangci-lint"
	@echo "  fmt        Run gofmt and goimports"
	@echo "  tidy       Run go mod tidy"
	@echo "  clean      Remove build artifacts"
