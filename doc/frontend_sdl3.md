# The SDL3 frontend

`a2sdl3` is [a2sdl](frontend_sdl2.md) on SDL3. It does the same things, with the
same keys, the same diskettes dropped on the window and the same joysticks. The
difference is in what it takes to build it: it uses SDL3 through
[go-sdl3](https://github.com/Zyko0/go-sdl3), which needs neither cgo, nor a C
compiler, nor SDL developer files. SDL3 itself travels inside the binary, and
is extracted to a temporary directory while the emulator runs.

It is experimental and it is not in the
[releases](https://github.com/ivanizag/izapple2/releases), which stay on SDL2.

## Building

On Linux, MacOS and Windows alike, with nothing installed but Go:

``` terminal
git clone github.com/ivanizag/izapple2
cd izapple2/frontend/a2sdl3
go build .
```

As there is no cgo involved it also cross compiles, so a single machine can
build for all of them:

``` terminal
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build .
```

The resulting binary carries the SDL3 library for the target platform, so there
is no DLL to copy next to it.

## Running

``` terminal
casa@servidor:~$ ./a2sdl3
```

The same command line as every other frontend, see
[command_line.md](command_line.md). Press F1 for the help.

## Keys

The same as [a2sdl](frontend_sdl2.md#keys), including the paste with
Shift-Insert or Cmd-V and the Open-Apple and Closed-Apple on the alt or option
keys.

The Apple keys are found by scancode instead of by keycode. On the keyboard
layouts where the right alt is AltGr, SDL3 does not report it as an alt key,
and the Closed-Apple would never be pressed if it were looked up by keycode.

## What is different from SDL2

Nothing that you use, and a few things underneath:

- SDL3 reports a file being dragged over the window, so the screen with the
  areas of the drives it can be dropped on is shown while it moves, with the
  one under the pointer marked. SDL2 only knows about the file once it is
  dropped, and has to show the areas with F8 or after the drop.
- Losing the window focus releases the joystick keys, so an Open-Apple held
  while switching windows does not stay pressed.
- Ctrl-C on the terminal quits the same way as closing the window, releasing
  everything on the way out.
- The scaling of the picture is a property of the renderer instead of the
  global hint SDL2 uses.
