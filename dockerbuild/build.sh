#!/bin/bash
cd "$( dirname $0)"
docker build . -t apple2builder --platform linux/amd64
mkdir -p ${PWD}/build
docker run --rm -it -v ${PWD}/build:/build apple2builder
