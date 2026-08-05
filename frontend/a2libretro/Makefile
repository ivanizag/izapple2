# Builds the libretro core for the host platform.
#
# The core is a C shared library, so a C toolchain is needed. Cross compiling
# needs the toolchain of the target platform and the matching CC, GOOS and
# GOARCH.

UNAME := $(shell uname -s)

ifeq ($(UNAME),Darwin)
	EXT := dylib
else ifneq (,$(findstring MINGW,$(UNAME)))
	EXT := dll
else
	EXT := so
endif

TARGET := izapple2_libretro.$(EXT)

.PHONY: all clean

all: $(TARGET)

$(TARGET):
	CGO_ENABLED=1 go build -buildmode=c-shared -o $(TARGET) .

clean:
	rm -f izapple2_libretro.so izapple2_libretro.dylib izapple2_libretro.dll
	rm -f izapple2_libretro.h
