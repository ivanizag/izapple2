# The Ebitengine frontend

`a2ebiten` opens a window with [Ebitengine](https://ebitengine.org/). It has
the screen with all its modes, the sound and the same function keys as
[a2sdl](frontend_sdl2.md), and it is missing the things around them: no
diskettes dropped on the window, no joysticks and no mouse.

## Building

Ebitengine needs no SDL, and on Linux and MacOS it needs a C compiler and the
graphics developer files. See
[the Ebitengine install page](https://ebitengine.org/en/documents/install.html)
for the list of packages of your distribution. On Windows it needs nothing.

``` terminal
git clone github.com/ivanizag/izapple2
cd izapple2/frontend/a2ebiten
go build .
```

## Running

``` terminal
casa@servidor:~$ ./a2ebiten
```

The [command line options](command_line.md) are the same as everywhere else,
and the diskettes have to go in the command line: this frontend has no way to
insert one afterwards. Press F1 for the help.

## Keys

The same as [a2sdl](frontend_sdl2.md#keys), with two differences:

- Ctrl-F5 shows a readout with the speed of the emulated machine and the frame
  rate on the corner of the screen, instead of printing the speed on the
  terminal.
- There is no paste from the clipboard.

The help screen that F1 shows is the one of a2sdl and mentions dropping a file
on the window. That does not work here.

## What is missing

- **Diskettes dropped on the window.** They go in the command line.
- **Joysticks and paddles**, and the mouse as a joystick.
- **The mouse**, so the models that use it, like `desktop`, are not much use.
- **Pasting** from the clipboard.

The window is 564x384 and can be resized. The picture is generated at a fixed
1128x768 and scaled to the window, so the modes wider than 560 pixels, like the
Videx cards or the RGB card, are squeezed to fit instead of widening the
window.
