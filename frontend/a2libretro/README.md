# izapple2 libretro core

A [libretro](https://www.libretro.com/) core, to run izapple2 inside RetroArch
and the other libretro frontends.

**If you want to use the core, read [doc/frontend_libretro.md](../../doc/frontend_libretro.md)**,
or [doc/frontend_libretro_macos.md](../../doc/frontend_libretro_macos.md) on a Mac,
which covers building it for each platform, installing it in the frontends,
loading disks, the controls and what it supports. What follows are the notes on
how it is put together.

## Build

```bash
make                 # izapple2_libretro.{so,dylib,dll} for the host
make universal       # macOS, one dylib with both architectures inside
```

Or directly:

```bash
go build -buildmode=c-shared -o izapple2_libretro.so ./frontend/a2libretro
```

`izapple2_libretro.info` goes with the library, in the info directory of the
frontend. Keep `corename`, `display_version` and `supported_extensions` in it
in step with what `retro_get_system_info` reports in `core.go`.

## How it differs from the other frontends

The other frontends let the emulation run free on its own goroutine and sample
the video and the audio whenever they need them. Here the frontend owns the
frame timing: it calls `retro_run()` once per frame, and the core has to
advance the machine by one frame, hand over the video frame and the matching
audio samples, and return. That is what `Apple2.RunCycles` is for, and it means
everything happens on the thread that calls `retro_run()`, with no goroutines
involved.

A frame is 17030 cycles, 65 for each of the 262 lines of an NTSC frame.

While a device asks for fast mode, which the Disk II does while the motor is
on, the core runs 16 frames worth of cycles per frame. Without it a DOS 3.3
boot, that spends some 37 million cycles mostly waiting for the drive, would
take more than half a minute. Sound is choppy while it lasts, as it is on the
other frontends when they run at full speed.

## What is not supported

Kept out on purpose, to keep the core small:

- **Super hi-res and the Videx 80 column video.** Every remaining mode is
  generated as 560x192, four pixels more with the NTSC color filter, so the
  framebuffer is a fixed 564x192 and the geometry never changes. The cards that
  generate the other modes are taken out of the slots, and the `ultraterm`
  model is not offered. The 80 columns of the Apple //e are not affected, they
  are a mode of the machine and land in the same 560x192 as the rest.
- **The RGB card**, that would widen the framebuffer to 644.
- **The gaps between the scan lines.** The core only uses the two screen modes
  without them, `screen.ScreenModeColor` and `screen.ScreenModeGreen`, so the
  image keeps its 192 lines instead of becoming four times taller, and the
  frontend draws the CRT look with a shader if it wants one. The phosphor is the
  other axis of the screen modes and is kept as a core option: a green monitor
  shows half width pixels that no shader can recover from a colour picture.
- **Savestates.** `retro_serialize_size()` reports zero, so the frontend
  disables rewind, run ahead and netplay. The emulator has no way to serialize
  the machine.
- **The mouse.** A stub provider is attached because the mouse card reads its
  provider without checking that there is one, so the models that carry it, like
  `desktop`, would crash the frontend.

## Where the games save

The emulation writes to a disk in the file it was loaded from, which is right
on the command line and wrong in a frontend that expects the content to be left
alone. `saves.go` asks the frontend for its save directory and passes it as the
`saveDir` configuration of the machine, the same option the command line has.
It is a setting of the machine, applied when it is built, so it is given as a
configuration override rather than set afterwards: the disks are inserted while
the machine is being built. From there `setupCard` passes it to the cards that
declare a `savedir` parameter and do not name one themselves.

The overlay itself is in the storage package, `overlay.go`, and is shared with
the other frontends. The core only chooses where it lives.

## Swapping diskettes

`disk.go` implements the extended disk control interface, falling back to the
older one without labels if the frontend does not know it. The list of
diskettes comes from an `.m3u` playlist; a single disk image gives a list of
one.

The drive has no eject, so ejecting only bookkeeps and the diskette is inserted
when the tray is closed again, which is the order the frontends use. Inserting
goes through `SendLoadDisk`, which queues on the command channel that
`RunCycles` drains at the top of the next frame, so it lands on the right
thread by itself. The selected diskette is also kept in `filenames`, so a reset
boots what is in the drive rather than what was loaded first.

## The C bridge

Go can not call the function pointers that the frontend hands to the core, so
`shim.c` keeps them and reaches them through plain C functions. The keyboard
callback goes the other way: the frontend needs a C function pointer, so
`shim_keyboard_callback` forwards to a Go function exported with cgo.

Four entry points, `retro_load_game`, `retro_load_game_special`,
`retro_unserialize` and `retro_cheat_set`, take a const pointer. cgo can not
put const in the declarations it generates for exported Go functions, and they
would clash with the ones in `libretro.h`, so those are exported under another
name and wrapped in `shim.c`.

Two more are in `shim.c` because they must not reach Go at all.
`retro_api_version` and `retro_get_system_info` are asked on the main thread of
the frontend on the way to loading the core, and every function exported from Go
begins by waiting for the Go runtime to finish starting, which finishes on that
same thread. Answering them from Go deadlocks the frontend for good. They are
constants, so they are answered in C.

### When the frontend may call into Go

The rule that came out of the two deadlocks: **the frontend must not reach Go
before there is a machine.** Everything it can call at any moment of its
choosing is either pure C or gated in C.

The keyboard callback is the interesting one. It is handed over in
`retro_set_environment`, early, because that is what tells RetroArch this is a
machine with a keyboard and lets its game focus come on by itself. But the
frontend then holds a pointer it may call from any thread while the core is
still loading, and going into Go there kills the process with a
`morestack on g0`. So `shim_keyboard_callback` drops the keys until
`shim_set_keyboard_ready` opens it, which `retro_load_game` does once the
machine is built and `core.close` closes again.

`libretro.h` is vendored from
[libretro-common](https://github.com/libretro/libretro-common).
