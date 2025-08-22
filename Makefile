# Makefile for redis-valkey-tui

VERSION ?= dev
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS = -X github.com/mohan-s-gopal/redis-valkey-tui/cmd.Version=$(VERSION) \
          -X github.com/mohan-s-gopal/redis-valkey-tui/cmd.BuildTime=$(BUILD_TIME) \
          -X github.com/mohan-s-gopal/redis-valkey-tui/cmd.GitCommit=$(GIT_COMMIT)

.PHONY: build run clean test install deps version build-all fmt mod tidy

# Build the application
build:
	go build -ldflags "$(LDFLAGS)" -o redis-valkey-tui .

# Run the application
run:
	go run .

# Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"

# Clean build artifacts
clean:
	rm -f redis-valkey-tui

# Run tests
test:
	go test ./internal/ui
	go test ./internal/ui/ -v -run TestInputHandling

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install the application to GOPATH/bin
install:
	go install .

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o redis-valkey-tui-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o redis-valkey-tui-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o redis-valkey-tui-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o redis-valkey-tui-windows-amd64.exe .

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  run        - Run the application"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  deps       - Install dependencies"
	@echo "  install    - Install to GOPATH/bin"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo "  help       - Show this help"
