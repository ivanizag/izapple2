# The Ebitengine frontend

`a2ebiten` opens a window with [Ebitengine](https://ebitengine.org/). It has
the screen with all its modes, the sound, the mouse, the diskettes dropped on
the window and the same function keys as [a2sdl](frontend_sdl2.md), and it is
missing the things around them: no joysticks.

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

The [command line options](command_line.md) are the same as everywhere else.
Press F1 for the help.

## Keys

The same as [a2sdl](frontend_sdl2.md#keys), with two differences:

- Ctrl-F5 shows a readout with the speed of the emulated machine and the frame
  rate on the corner of the screen, instead of printing the speed on the
  terminal.
- There is no paste from the clipboard.

## Diskettes

Drop a file on the window to insert it. The window is divided in as many
vertical areas as removable media drives the machine has, and the file goes to
the drive of the area it is dropped on. F8 shows a screen with the areas, each
one with the name of its drive and the image it has inserted, or `EMPTY`, and
it is shown again for a moment after a drop with the drive that got the file
marked.

Ebitengine hands over the files dropped as a file system that hides their
paths, but it opens the real files and the handle of a file tells its path
back. The image is loaded from there like on any other frontend, compressed
files included, and whatever the emulated software writes goes back to the file
or to the `-saveDir` directory. In the browser the files dropped have no path
and cannot be loaded.

Ebitengine does not report where a file was dropped either, and the pointer
position is not updated while another application drags a file over the window,
so the area used is the one the pointer was last seen on. Check with F8 before
dragging, or look at the areas shown after the drop to see where the file
landed.

## The mouse

The models that use a mouse, like `desktop`, work with the pointer on the
window. Ebitengine has no mouse events, so the pointer is read on every frame,
and its position is taken on the picture and not on the window: it lands on the
same place whatever the size of the window is.

## What is missing

- **Joysticks and paddles**, and the mouse as a joystick.
- **Pasting** from the clipboard.

The window is 564x384 and can be resized. The picture is generated at a fixed
1128x768 and scaled to the window, so the modes wider than 560 pixels, like the
Videx cards or the RGB card, are squeezed to fit instead of widening the
window.
