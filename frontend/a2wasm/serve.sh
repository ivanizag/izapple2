#!/bin/bash
# Development server for izapple2 WebAssembly frontend

set -e

PORT=8080

echo "========================================="
echo "  izapple2 WebAssembly Dev Server"
echo "========================================="
echo ""

# Check if WASM file exists
if [ ! -f "web/izapple2.wasm" ]; then
    echo "Error: web/izapple2.wasm not found"
    echo ""
    echo "Please run ./build.sh first"
    exit 1
fi

echo "Starting HTTP server on port $PORT..."
echo ""
echo "Open in browser:"
echo "  http://localhost:$PORT"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Try different server options in order of preference
if command -v python3 &> /dev/null; then
    # Python 3 HTTP server
    echo "Using Python 3 HTTP server"
    cd web
    python3 -m http.server $PORT
elif command -v python &> /dev/null; then
    # Python 2 HTTP server (fallback)
    echo "Using Python 2 HTTP server"
    cd web
    python -m SimpleHTTPServer $PORT
elif command -v go &> /dev/null; then
    # Use Go's wasmserve if available
    if go list github.com/hajimehoshi/wasmserve &> /dev/null; then
        echo "Using wasmserve"
        go run github.com/hajimehoshi/wasmserve@latest .
    else
        echo "Error: No suitable HTTP server found"
        echo ""
        echo "Please install one of:"
        echo "  - Python 3 (recommended)"
        echo "  - Python 2"
        echo "  - Or run: go install github.com/hajimehoshi/wasmserve@latest"
        exit 1
    fi
else
    echo "Error: No suitable HTTP server found"
    echo ""
    echo "Please install Python 3 or Go"
    exit 1
fi
