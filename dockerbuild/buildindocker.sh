#!/bin/bash
cd /tmp
git clone https://github.com/ivanizag/apple2

# Build apple2console for Linux
cd /tmp/apple2/apple2console
go build .
chown --reference /build apple2console
cp apple2console /build

# Build apple2console.exe for Windows
cd /tmp/apple2/apple2console
env CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows CGO_LDFLAGS="-L/usr/x86_64-w64-mingw32/lib" CGO_FLAGS="-I/usr/x86_64-w64-mingw32/include -D_REENTRANT" go build -o apple2console.exe .
chown --reference /build apple2console.exe
cp apple2console.exe /build

# Build apple2sdl for Linux
cd /tmp/apple2/apple2sdl
go build .
chown --reference /build apple2sdl
cp apple2sdl /build

# Build apple2sdl.exe for Windows
cd /tmp/apple2/apple2sdl
env CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows CGO_LDFLAGS="-L/usr/x86_64-w64-mingw32/lib -lSDL2" CGO_FLAGS="-I/usr/x86_64-w64-mingw32/include -D_REENTRANT" go build -o apple2sdl.exe .
chown --reference /build apple2sdl.exe
cp apple2sdl.exe /build
