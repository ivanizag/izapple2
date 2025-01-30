#!/bin/bash
cd "$( dirname $0)"
mkdir -p ${PWD}/build

# MacOS ARM builds
echo "Building MacOS ARM console frontend"
CGO_ENABLED=1 go build -tags static -ldflags "-s -w" ../frontend/console
mv console build/izapple2console_mac_arm64

echo "Building MacOS ARM SDL frontend"
CGO_ENABLED=1 go build -tags static -ldflags "-s -w" ../frontend/a2sdl
mv a2sdl build/izapple2sdl_mac_arm64

#echo "Building MacOS Fyne frontend"
#CGO_ENABLED=1 go build -tags static -ldflags "-s -w" ../frontend/a2fyne
#mv a2fyne build/izapple2fyne_mac_arm64

# MacOS x64 builds
echo "Building MacOS Intel console frontend"
GOARCH=amd64 CGO_ENABLED=1 go build -tags static -ldflags "-s -w" ../frontend/console
mv console build/izapple2console_mac_amd64

echo "Building MacOS Intel SDL frontend"
GOARCH=amd64 CGO_ENABLED=1 go build -tags static -ldflags "-s -w" ../frontend/a2sdl
mv a2sdl build/izapple2sdl_mac_amd64

#echo "Building MacOS Fyne frontend"
#GOARCH=amd64 CGO_ENABLED=1 go build -tags static -ldflags "-s -w" ../frontend/a2fyne
#mv a2fyne build/izapple2fyne_mac_amd64

# Linux and Windows dockerized builds
echo "Building docker container for the Linux and Windows builds"
docker build . -t apple2builder --platform linux/amd64
docker run --rm -it -v ${PWD}/build:/build apple2builder

cd build
cp ../../README.md .
zip izapple2sdl_windows_amd64.zip izapple2sdl_windows_amd64.exe README-SDL.txt README.md SDL2.dll 


