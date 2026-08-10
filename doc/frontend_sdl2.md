# The SDL2 frontend

`a2sdl` is the complete frontend, the one the
[releases](https://github.com/ivanizag/izapple2/releases) and the Homebrew
formula contain. It opens a window with SDL2 and gives you everything the
emulator has: the screen with all its modes, sound, joysticks, the mouse,
diskettes dropped on the window and the whole set of function keys.

Use this one unless you have a reason not to. The reason is usually SDL2
itself, which needs a C compiler and its developer files to build: if that is
in the way, [a2sdl3](frontend_sdl3.md) is the same frontend without any of it.

## Building

Besides a working Go installation you need a C compiler and the SDL2 developer
files.

### Linux

``` terminal
sudo apt-get install libsdl2-dev
git clone github.com/ivanizag/izapple2
cd izapple2/frontend/a2sdl
go build .
```

### MacOS

``` terminal
brew install SDL2
git clone github.com/ivanizag/izapple2
cd izapple2/frontend/a2sdl
go build .
```

### Windows

CGO needs a gcc compiler. Install
[mingw-w64](http://mingw-w64.org/doku.php/download/mingw-builds) and the
[SDL2 developer files](https://www.libsdl.org/release/) for mingw-64, and run:

``` terminal
git clone github.com/ivanizag/izapple2
cd izapple2\frontend\a2sdl
go build .
```

To run it, copy `SDL2.dll` to the same folder as `a2sdl.exe`. The latest one is
in the [Runtime binary for Windows 64-bit](https://www.libsdl.org/download-2.0.php).

### A single file executable

The released binaries are built with SDL2 linked statically, so that they run
on a machine without SDL2 installed:

``` terminal
go build -tags static -ldflags "-s -w" -o izapple2 .
```

That is what the release workflow does, see [release.md](release.md).

## Running

``` terminal
casa@servidor:~$ ./a2sdl
```

Without arguments you get the default machine, an enhanced Apple //e booting
DOS 3.3. A file on the command line is loaded in the right place by its
extension, and everything else is set with the
[command line options](command_line.md):

``` terminal
casa@servidor:~$ ./a2sdl "https://www.apple.asimov.net/images/games/action/karateka/karateka (includes intro).dsk"
casa@servidor:~$ ./a2sdl -model 2plus Total\ Replay\ v6.0.1.hdv
```

The window is resizable and the picture follows it. Press F1 for the list of
keys without leaving the emulator.

## Keys

The Apple II keyboard is where you expect it, with the arrows, Tab, Delete and
Escape mapped to what the Apple II understands. The rest:

| Key | What it does |
| --- | --- |
| F1 | Show or hide the help |
| Ctrl-F2 | Reset. F2 alone also resets while the help is shown, for the window managers that eat Ctrl-F2 |
| F4 | Show or hide the CPU trace on the terminal |
| F5 | Full speed or normal speed |
| Ctrl-F5 | Print the current speed on the terminal |
| F6 | Next screen mode: NTSC colour, plain, green, with or without scan lines |
| F7 | Show or hide the four panels with the actual screen, page 1, page 2 and the extra info of the video mode |
| F8 | Show or hide the areas where a diskette can be dropped |
| F9 | Dump the state of the machine on the terminal |
| F10 | Next character set |
| Ctrl-F10 | Show or hide the character map |
| Shift-F10 | Show or hide the alternate text of the character map |
| F12 or PrintScreen | Save the screen as `snapshot.png` |
| Pause | Pause and unpause the emulation |
| Left alt or option | Open-Apple |
| Right alt or option | Closed-Apple |
| Shift-Insert, Cmd-V | Paste the clipboard, typed in slowly so that the machine keeps up |

On the Base64A, F3 is the Delete key of that keyboard.

The Apple keys are a place on the keyboard and are found by scancode, so they
work on the layouts where the alt keys produce another symbol. Losing the
window focus releases them, so one held while switching windows does not stay
pressed.

The title bar says which machine is running and shows `PAUSED!` while it is
paused. With the four panels or the character map on screen it shows what is
being displayed instead.

## Diskettes

Drop a file on the window to insert it. The window is divided in as many
vertical areas as removable media drives the machine has, and the file goes to
the drive of the area it is dropped on. It works with everything that goes on
the command line, including compressed images.

F8 shows a screen with the areas, each one with the name of its drive and the
image it has inserted, or `EMPTY`, and the one under the pointer marked. It
takes over the picture the same way the help does, on 80 columns. The same
screen is shown for a moment after a drop, with the drive that got the file
marked.

SDL2 does not report a file being dragged over the window, so the areas cannot
be shown while the file moves; the position of the pointer is read when the
file is dropped. The [SDL3 frontend](frontend_sdl3.md) does show them during
the drag. The areas are the same on every frontend, they are built in
`frontend/shared`.

Whatever the emulated software writes goes back to the file it came from. To
keep the images untouched use `-saveDir`, which puts the writes in an overlay
file per disk and makes writable even the disks loaded from an URL, a
compressed file or the embedded resources.

## Joysticks and mouse

Joysticks connected to the host are used as the Apple II joysticks, up to two
of them or four paddles. If the machine has no mouse card, the host mouse acts
as a joystick too.

The mouse card needs no capture: the pointer of the host is the pointer of the
Apple II, wherever it is on the window.

## Sound

The speaker and the sound cards, like the Mockingboard, are mixed and sent to
the SDL2 audio device. At full speed the sound gets choppy, which is what
happens when the machine that generates it runs faster than the sound it
generates.
