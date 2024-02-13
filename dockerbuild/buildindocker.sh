#!/bin/bash
cd /tmp
git clone https://github.com/ivanizag/izapple2

# Build izapple2console for Linux
echo "Building Linux console frontend"
cd /tmp/izapple2/frontend/console
env CGO_ENABLED=1 go build -tags static -ldflags "-s -w" .
chown --reference /build console
cp console /build/izapple2console_linux_amd64

# Build izapple2console.exe for Windows
echo "Building Windows console frontend"
cd /tmp/izapple2/frontend/console
env CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows CGO_LDFLAGS="-L/usr/x86_64-w64-mingw32/lib" CGO_FLAGS="-I/usr/x86_64-w64-mingw32/include -D_REENTRANT" go build -tags static -ldflags "-s -w" -o izapple2console.exe .
chown --reference /build izapple2console.exe
cp izapple2console.exe /build/izapple2console_windows_amd64.exe

# Build izapple2sdl for Linux
echo "Building Linux SDL frontend"
cd /tmp/izapple2/frontend/a2sdl
env CGO_ENABLED=1 go build -tags static -ldflags "-s -w" .
chown --reference /build a2sdl
cp a2sdl /build/izapple2sdl_linux_amd64

# Build izapple2sdl.exe for Windows
echo "Building Windows SDL frontend"
cd /tmp/izapple2/frontend/a2sdl
env CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows CGO_LDFLAGS="-L/usr/x86_64-w64-mingw32/lib -lSDL2" CGO_FLAGS="-I/usr/x86_64-w64-mingw32/include -D_REENTRANT" go build -tags static -ldflags "-s -w" -o izapple2sdl.exe .
chown --reference /build izapple2sdl.exe
cp izapple2sdl.exe /build/izapple2sdl_windows_amd64.exe

# Build izapple2fyne for Linux
echo "Building Linux Fyne frontend"
cd /tmp/izapple2/frontend/a2fyne
env CGO_ENABLED=1 go build -tags static -ldflags "-s -w" .
chown --reference /build a2fyne
cp a2fyne /build/izapple2fyne_linux_amd64

# Build izapple2fyne.exe for Windows
echo "Building Windows Fyne frontend"
cd /tmp/izapple2/frontend/a2fyne
env CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows CGO_LDFLAGS="-L/usr/x86_64-w64-mingw32/lib " CGO_FLAGS="-I/usr/x86_64-w64-mingw32/include -D_REENTRANT" go build -tags static -ldflags "-s -w" -o izapple2fyne.exe .
chown --reference /build izapple2fyne.exe
cp izapple2fyne.exe /build/izapple2fyne_windows_amd64.exe


# Copy SDL2 Runtime
cp /sdl2runtime/* /build
