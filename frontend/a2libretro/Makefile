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

.PHONY: all universal clean

all: $(TARGET)

$(TARGET):
	CGO_ENABLED=1 go build -buildmode=c-shared -o $(TARGET) .

# One core for both architectures of the Mac. It is worth having because a
# frontend can only load a core of its own architecture, and the RetroArch that
# Homebrew installs is an Intel build that runs under Rosetta even on an Apple
# Silicon Mac, where the native core would not load.
universal:
ifneq ($(UNAME),Darwin)
	$(error universal is only for macOS)
endif
	CGO_ENABLED=1 GOARCH=arm64 CC="clang -arch arm64" \
		go build -buildmode=c-shared -o izapple2_libretro_arm64.dylib .
	CGO_ENABLED=1 GOARCH=amd64 CC="clang -arch x86_64" \
		go build -buildmode=c-shared -o izapple2_libretro_x86_64.dylib .
	lipo -create izapple2_libretro_arm64.dylib izapple2_libretro_x86_64.dylib \
		-output izapple2_libretro.dylib
	rm -f izapple2_libretro_arm64.dylib izapple2_libretro_x86_64.dylib
	rm -f izapple2_libretro_arm64.h izapple2_libretro_x86_64.h
	@lipo -archs izapple2_libretro.dylib

clean:
	rm -f izapple2_libretro.so izapple2_libretro.dylib izapple2_libretro.dll
	rm -f izapple2_libretro.h
	rm -f izapple2_libretro_arm64.dylib izapple2_libretro_x86_64.dylib
	rm -f izapple2_libretro_arm64.h izapple2_libretro_x86_64.h
