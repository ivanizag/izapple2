#!/bin/bash
# Build script for izapple2 WebAssembly + React frontend

set -e

echo "========================================="
echo "  izapple2 WASM + React Build Script"
echo "========================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "Go version: $GO_VERSION"

# Get GOROOT for wasm_exec.js
GOROOT=$(go env GOROOT)
WASM_EXEC_JS="$GOROOT/misc/wasm/wasm_exec.js"

# Check alternate locations
if [ ! -f "$WASM_EXEC_JS" ]; then
    WASM_EXEC_JS="$GOROOT/lib/wasm/wasm_exec.js"
fi
if [ ! -f "$WASM_EXEC_JS" ]; then
    WASM_EXEC_JS="/usr/local/go/misc/wasm/wasm_exec.js"
fi
if [ ! -f "$WASM_EXEC_JS" ]; then
    WASM_EXEC_JS="/opt/homebrew/opt/go/libexec/misc/wasm/wasm_exec.js"
fi

if [ ! -f "$WASM_EXEC_JS" ]; then
    echo "Error: Could not find wasm_exec.js"
    exit 1
fi

echo ""
echo "Building Go to WebAssembly..."
echo "Command: cd go && GOOS=js GOARCH=wasm go build -ldflags=\"-s -w\" -o ../public/wasm/izapple2.wasm"
echo ""

# Build WASM binary
cd go
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o ../public/wasm/izapple2.wasm

if [ $? -ne 0 ]; then
    echo ""
    echo "Error: WASM build failed"
    exit 1
fi

cd ..

# Get file size
WASM_SIZE=$(ls -lh public/wasm/izapple2.wasm | awk '{print $5}')
echo ""
echo "✓ WASM build successful!"
echo "  Output: public/wasm/izapple2.wasm ($WASM_SIZE)"
echo ""

# Copy wasm_exec.js
echo "Copying wasm_exec.js..."
cp "$WASM_EXEC_JS" public/wasm/wasm_exec.js

if [ $? -ne 0 ]; then
    echo "Error: Failed to copy wasm_exec.js"
    exit 1
fi

echo "✓ Copied wasm_exec.js"
echo ""

# Optional: Compress with gzip
if command -v gzip &> /dev/null; then
    echo "Compressing WASM binary with gzip..."
    gzip -k -f public/wasm/izapple2.wasm
    GZIP_SIZE=$(ls -lh public/wasm/izapple2.wasm.gz | awk '{print $5}')
    echo "✓ Created public/wasm/izapple2.wasm.gz ($GZIP_SIZE)"
    echo ""
fi

echo "========================================="
echo "  WASM Build Complete!"
echo "========================================="
echo ""
echo "To run React dev server:"
echo "  npm run dev"
echo ""
echo "To build for production:"
echo "  npm run build"
echo ""
