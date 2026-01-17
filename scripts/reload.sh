#!/bin/bash

# TRMNL Renderer reload script
# Rebuilds and restarts the service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "Building TRMNL Renderer..."

# Download dependencies if needed
if [ ! -f "go.sum" ]; then
    echo "Downloading dependencies..."
    go mod download
fi

# Build Go binary
echo "Building Go binary..."
go build -o trmnl-renderer main.go render.go server.go styles.go views.go rotation.go

echo "Build complete!"
echo ""
echo "To start the service, run: ./trmnl-renderer"
echo "Or if using a process manager, restart the service now."

