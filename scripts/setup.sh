#!/bin/bash

# Setup script for redis-cli-dashboard development

set -e

echo "Setting up redis-cli-dashboard development environment..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | cut -d'.' -f1,2)
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION or later."
    exit 1
fi

echo "Go version: $GO_VERSION ✓"

# Install dependencies
echo "Installing dependencies..."
go mod download

# Build the application
echo "Building redis-cli-dashboard..."
make build

# Run tests
echo "Running tests..."
go test -v .

# Create example config
echo "Creating example config..."
mkdir -p ~/.redis-cli-dashboard
if [ ! -f ~/.redis-cli-dashboard/config.json ]; then
    cp config.example.json ~/.redis-cli-dashboard/config.json
    echo "Example config created at ~/.redis-cli-dashboard/config.json"
fi

echo ""
echo "Setup complete! 🎉"
echo ""
echo "To get started:"
echo "  ./redis-cli-dashboard -help    # Show help"
echo "  ./redis-cli-dashboard          # Connect to localhost:6379"
echo "  make run          # Run with make"
echo ""
echo "For more information, see README.md and USAGE.md"
