#!/bin/bash
cd /tmp
git clone https://github.com/ivanizag/izapple2

# Build izapple2console for Linux
cd /tmp/izapple2/izapple2console
go build .
chown --reference /build izapple2console
cp izapple2console /build

# Build izapple2console.exe for Windows
cd /tmp/izapple2/izapple2console
env CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows CGO_LDFLAGS="-L/usr/x86_64-w64-mingw32/lib" CGO_FLAGS="-I/usr/x86_64-w64-mingw32/include -D_REENTRANT" go build -o izapple2console.exe .
chown --reference /build izapple2console.exe
cp izapple2console.exe /build

# Build izapple2sdl for Linux
cd /tmp/izapple2/izapple2sdl
go build .
chown --reference /build izapple2sdl
cp izapple2sdl /build

# Build izapple2sdl.exe for Windows
cd /tmp/izapple2/izapple2sdl
env CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows CGO_LDFLAGS="-L/usr/x86_64-w64-mingw32/lib -lSDL2" CGO_FLAGS="-I/usr/x86_64-w64-mingw32/include -D_REENTRANT" go build -o izapple2sdl.exe .
chown --reference /build izapple2sdl.exe
cp izapple2sdl.exe /build

# Copy SDL2 Runtime
cp /sdl2runtime/* /build
