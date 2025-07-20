# Makefile for redis-cli-dashboard

.PHONY: build run clean test install deps

# Build the application
build:
	go build -o redis-cli-dashboard .

# Run the application
run:
	go run .

# Clean build artifacts
clean:
	rm -f redis-cli-dashboard

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
	GOOS=linux GOARCH=amd64 go build -o redis-cli-dashboard-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o redis-cli-dashboard-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o redis-cli-dashboard-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o redis-cli-dashboard-windows-amd64.exe .

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
