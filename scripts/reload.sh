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

# Build Go binary (build tags will automatically select correct tray implementation)
echo "Building Go binary..."
# On Linux, tray_noop.go will be included (build tag !windows)
# On Windows, tray.go will be included (build tag windows)
go build -o trmnl-renderer ./main.go ./render.go ./server.go ./styles.go ./views.go ./rotation.go ./tray_noop.go

echo "Build complete!"
echo ""
echo "To start the service, run: ./trmnl-renderer"
echo "Or if using a process manager, restart the service now."

