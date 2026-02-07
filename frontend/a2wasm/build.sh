#!/bin/bash
# Build script for izapple2 WebAssembly frontend

set -e

echo "========================================="
echo "  izapple2 WebAssembly Build Script"
echo "========================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Check Go version (need 1.16+ for embed)
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "Go version: $GO_VERSION"

# Get GOROOT for wasm_exec.js
GOROOT=$(go env GOROOT)
WASM_EXEC_JS="$GOROOT/misc/wasm/wasm_exec.js"

# Check if wasm_exec.js exists (try alternate locations)
if [ ! -f "$WASM_EXEC_JS" ]; then
    WASM_EXEC_JS="$GOROOT/lib/wasm/wasm_exec.js"
fi

if [ ! -f "$WASM_EXEC_JS" ]; then
    # Try standard installation location
    WASM_EXEC_JS="/usr/local/go/misc/wasm/wasm_exec.js"
fi

if [ ! -f "$WASM_EXEC_JS" ]; then
    # Try homebrew location on macOS
    WASM_EXEC_JS="/opt/homebrew/opt/go/libexec/misc/wasm/wasm_exec.js"
fi

if [ ! -f "$WASM_EXEC_JS" ]; then
    echo "Error: Could not find wasm_exec.js in Go installation"
    echo "Searched:"
    echo "  - $GOROOT/misc/wasm/wasm_exec.js"
    echo "  - $GOROOT/lib/wasm/wasm_exec.js"
    echo "  - /usr/local/go/misc/wasm/wasm_exec.js"
    echo "  - /opt/homebrew/opt/go/libexec/misc/wasm/wasm_exec.js"
    exit 1
fi

echo "GOROOT: $GOROOT"
echo ""

# Build WASM binary
echo "Compiling Go to WebAssembly..."
echo "Command: GOOS=js GOARCH=wasm go build -ldflags=\"-s -w\" -o web/izapple2.wasm"
echo ""

GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/izapple2.wasm

if [ $? -ne 0 ]; then
    echo ""
    echo "Error: Build failed"
    exit 1
fi

# Get file size
WASM_SIZE=$(ls -lh web/izapple2.wasm | awk '{print $5}')
echo ""
echo "✓ Build successful!"
echo "  Output: web/izapple2.wasm ($WASM_SIZE)"
echo ""

# Copy wasm_exec.js
echo "Copying wasm_exec.js from Go installation..."
cp "$WASM_EXEC_JS" web/wasm_exec.js

if [ $? -ne 0 ]; then
    echo "Error: Failed to copy wasm_exec.js"
    exit 1
fi

echo "✓ Copied wasm_exec.js"
echo ""

# Optional: Compress with gzip
if command -v gzip &> /dev/null; then
    echo "Compressing WASM binary with gzip..."
    gzip -k -f web/izapple2.wasm
    GZIP_SIZE=$(ls -lh web/izapple2.wasm.gz | awk '{print $5}')
    echo "✓ Created web/izapple2.wasm.gz ($GZIP_SIZE)"
    echo ""
fi

# Display file listing
echo "Build output files:"
ls -lh web/ | grep -E '\.(html|js|css|wasm)' || true
echo ""

echo "========================================="
echo "  Build Complete!"
echo "========================================="
echo ""
echo "To test locally:"
echo "  ./serve.sh"
echo ""
echo "Or manually:"
echo "  cd web"
echo "  python3 -m http.server 8080"
echo ""
echo "Then open: http://localhost:8080"
echo ""
