# The Fyne frontend

`a2fyne` opens a window with [Fyne](https://fyne.io/). Unlike the other
frontends with a window, it has a proper interface around the screen: a toolbar
with the controls, and a side panel listing the cards in the slots with what
each of them is doing.

It is missing the latest features. Not much more is done there.

## Building

Fyne needs a C compiler and the graphics developer files of the platform. See
[the Fyne prerequisites](https://docs.fyne.io/started/) for the list of
packages of your distribution.

``` terminal
git clone github.com/ivanizag/izapple2
cd izapple2/frontend/a2fyne
go build .
```

## Running

``` terminal
casa@servidor:~$ ./a2fyne
```

The [command line options](command_line.md) are the same as everywhere else,
and the diskettes have to go in the command line: there is no way to insert one
afterwards.

## The toolbar

From left to right:

- Reset.
- Pause and unpause.
- Full speed or normal speed.
- The screen mode: NTSC colour, plain or green, the current one shown greyed
  out.
- Show or hide the four panels with the actual screen, page 1, page 2 and the
  extra info of the video mode.
- Save the screen as `snapshot.png`. It says so with a notification of the
  desktop.
- Force the keyboard to uppercase, for the software that expects a machine
  without lowercase. The title goes uppercase while it is on.
- Full screen.
- Show or hide the panel with the devices.

## The devices panel

On the right, and hidden until the toolbar asks for it. It has the state of the
joysticks and then a card per slot, with the name of the card and whatever it
reports about itself: the diskette in each drive of a Disk II, the tracks it is
on, and so on for the rest.

## Keys

Fewer than in the other frontends, and not the same ones:

| Key | What it does |
| --- | --- |
| Ctrl-F1 | Reset |
| F5 | Print the current speed on the terminal |
| F6 | Next screen mode |
| F7 | Show or hide the four panels |
| F9 | Dump the state of the machine on the terminal |
| F10 | Next character set |
| F11 | Show or hide the CPU trace on the terminal |
| F12 | Save the screen as `snapshot.png` |

Ctrl and a letter reach the machine as the control character. The arrows,
Return, Escape, Backspace and Delete are mapped as usual. On the Base64A, F2 is
the Delete key of that keyboard.

There is no Pause key and no PrintScreen, use the toolbar. The Tab key never
arrives.

## What is missing

- **Sound.** Nothing is connected to the speaker here.
- **The mouse**, so the models that use it, like `desktop`, are not much use.
- **Diskettes dropped on the window**, and the toolbar button to insert one is
  written but not enabled.
- **The character map** and the alternate text, which the other frontends show
  with F10.
- **A help**, the F1 of the other frontends.

Joysticks do work, read through GLFW.
