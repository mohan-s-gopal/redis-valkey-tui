#!/bin/bash

# Setup script for redis-valkey-tui development

set -e

echo "Setting up redis-valkey-tui development environment..."

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
echo "Building redis-valkey-tui..."
make build

# Run tests
echo "Running tests..."
go test -v .

# Create example config
echo "Creating example config..."
mkdir -p ~/.redis-valkey-tui
if [ ! -f ~/.redis-valkey-tui/config.json ]; then
    cp config.example.json ~/.redis-valkey-tui/config.json
    echo "Example config created at ~/.redis-valkey-tui/config.json"
fi

echo ""
echo "Setup complete! 🎉"
echo ""
echo "To get started:"
echo "  ./redis-valkey-tui -help    # Show help"
echo "  ./redis-valkey-tui          # Connect to localhost:6379"
echo "  make run          # Run with make"
echo ""
echo "For more information, see README.md and USAGE.md"
