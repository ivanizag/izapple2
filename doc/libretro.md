# Using izapple2 as a libretro core

izapple2 can be built as a [libretro](https://www.libretro.com/) core, so the
emulator runs inside RetroArch and the other libretro frontends instead of in
its own window. Your Apple II disks then sit in the same library as the rest of
your collection, with the same controller setup, video filters and shaders.

The core is not part of the [releases](https://github.com/ivanizag/izapple2/releases),
you have to build it from source. It is a small build, the instructions are below.

If you just want to run Apple II software on your computer, the regular
`izapple2sdl` executable is easier and does more. Use the core if you already
live in RetroArch.

On a Mac, [libretro_macos.md](libretro_macos.md) is the whole thing end to end,
from installing RetroArch to checking that it works. Follow that instead, there
are two traps on macOS that it walks you around.

## Building

The core is a shared library built with cgo, so besides Go you need a C
compiler. You do not need SDL2.

### Linux

``` terminal
git clone https://github.com/ivanizag/izapple2
cd izapple2/frontend/a2libretro
make
```

This produces `izapple2_libretro.so`.

### macOS

``` terminal
git clone https://github.com/ivanizag/izapple2
cd izapple2/frontend/a2libretro
make universal
```

This produces `izapple2_libretro.dylib` for both architectures of the Mac, so
that it loads in an Intel frontend and in an Apple Silicon one.

A frontend can only load a core of its own architecture, and the frontend has to
be running natively: a Go core cannot run inside a frontend translated by
Rosetta, whichever slice it loads. On macOS that rules out the `retroarch` cask
of Homebrew, which is an Intel build. See
[libretro_macos.md](libretro_macos.md).

### Windows

CGO needs a gcc compiler. Install
[mingw-w64](http://mingw-w64.org/doku.php/download/mingw-builds) and run:

``` terminal
git clone https://github.com/ivanizag/izapple2
cd izapple2\frontend\a2libretro
go build -buildmode=c-shared -o izapple2_libretro.dll .
```

## Installing the core

Copy the library you built into the cores directory of your frontend.

### RetroArch

| System | Cores directory |
| --- | --- |
| Linux | `~/.config/retroarch/cores` |
| macOS | `~/Library/Application Support/RetroArch/cores` |
| Windows | `cores` inside the RetroArch folder |

If you are not sure, RetroArch shows the path in
*Settings > Directory > Cores*.

Also copy `izapple2_libretro.info`, next to the sources in
`frontend/a2libretro`, into the info directory, shown in
*Settings > Directory > Core Info*. It is what makes RetroArch show the core
name, the supported extensions and the list of what it can and cannot do
instead of treating it as an unknown core. The core works without it, but the
content dialog will not filter by extension.

Then load it with *Load Core > Install or Restore a Core*, or pick it from
*Load Core* after restarting RetroArch.

The core is not in the RetroArch online updater, so it will not be downloaded
or updated for you, and the playlist scanner will not recognize Apple II disks
by itself. Load them by hand with *Load Core* followed by *Load Content*.

### Lakka, Batocera, RetroPie, Recalbox

All of these run RetroArch underneath, so the core works the same way. Copy the
`.so` you built for that machine's architecture into its cores directory
(`/tmp/cores` or `/usr/lib/libretro` on Lakka,
`/userdata/system/configs/retroarch/cores` on Batocera,
`/opt/retropie/libretrocores` on RetroPie) and select it as the core for your
Apple II games.

Be careful to build for the right architecture. A core built on your x86 laptop
will not load on a Raspberry Pi; you need to cross compile with the toolchain of
the target, setting `CC`, `GOOS` and `GOARCH`.

### Other frontends

Anything that loads libretro cores works: Kodi's game add-ons, Emulation
Station DE, Provenance, and the rest. Nothing in the core is specific to
RetroArch.

## Running software

### With no disk

Load the core and start it with no content. You get an enhanced Apple //e
booting DOS 3.3, ready for Applesoft BASIC. The ROMs and the DOS 3.3 diskette
are built into the core, there is nothing else to download.

In RetroArch this is *Load Core*, then *Start Core*.

### With a disk

Use *Load Content* and pick your disk image. These extensions are recognized:

| Extension | What it is |
| --- | --- |
| `dsk`, `do`, `po`, `nib`, `woz` | 5 1/4 diskettes |
| `2mg`, `hdv` | 3.5 disks and hard disks |
| `wav` | cassette tape recordings |
| `zip`, `gz` | any of the above, compressed |
| `m3u` | a playlist for a game on several diskettes, see below |

The core decides where to put the file the same way the command line does: a
diskette goes into the Disk II in slot 6, a hard disk image into a SmartPort
card, and a WAV into the cassette input.

### Games on several diskettes

Plenty of Apple II games ask you to insert disk 2 partway through. List the
diskettes in an `.m3u` playlist, a plain text file with one image per line, and
load that instead of a single disk:

``` terminal
Bard's Tale - Disk 1, Side A.woz
Bard's Tale - Disk 1, Side B.woz
Bard's Tale - Disk 2, Side A.woz
```

Relative paths are taken from the folder of the playlist, and lines starting
with `#` are ignored. The disks are named in the menu after their file names.

When the game asks for another disk, open *Quick Menu > Disk Control*, eject
the current disk, pick the one you need and insert it. RetroArch remembers
which one you were on the next time you load the playlist.

Only the first drive of the controller in slot 6 is swapped, which is what
games ask for.

### Saved games

Games that save write to the disk, so the core keeps your collection out of it.
The changes never reach the image you loaded: they go to an overlay file in the
save directory of the frontend, shown in *Settings > Directory > Saves* and
`~/Documents/RetroArch/saves` by default on macOS, named after the disk with
`.ovl` added.

An overlay only holds the parts the game actually wrote, not a copy of the
disk, so it is small: a few kilobytes for a diskette, and for a 32 Mb hard disk
only the blocks that were saved. Loading the game again puts them back on top
of the image and you are where you left off.

Because the image is only ever read, this also works for the disks that could
not be written at all before: the ones inside a zip or a gzip, the ones loaded
from an URL, and the DOS 3.3 built into the core. You can save to all of them
now.

Diskettes and hard disks both work this way, so a ProDOS collection on a HDV or
2MG is left untouched as well.

An overlay carries a checksum of the image it was made from, so it is never
applied to a different disk that happens to have the same name. If you get an
error saying the overlay belongs to another image, delete the `.ovl` file.

To start over from a clean disk, delete its `.ovl` file from the save
directory.

WOZ and NIB diskettes are read only by nature, and games on them cannot save.

This is not specific to the core. Any izapple2 frontend gets the same with the
`-saveDir` command line option, and each disk card can name a directory of its
own, see
[Keeping the Disk Images Unmodified](command_line.md#keeping-the-disk-images-unmodified).

### Speed

Disk access automatically runs faster than the real machine, as it does in the
other izapple2 frontends, so a DOS 3.3 boot takes about two seconds instead of
half a minute. The sound breaks up while this happens. That is expected.

RetroArch's own fast forward and slow motion work as usual on top of that.

## Core options

In *Quick Menu > Core Options*:

| Option | Values | Meaning |
| --- | --- | --- |
| Machine model | `gaming` (default), `2`, `2e`, `2enh`, `2plus`, `base64a`, `basis108`, `cpm`, `cpm3`, `cpm65`, `desktop`, `dos32`, `pascal`, `prodos`, `swyft` | The machine to emulate, the same models as the `-model` command line flag |
| Monitor | `color` (default), `green` | An NTSC colour television or a green phosphor monitor |

Changing the model takes effect on the next **Reset**, which cold boots the
machine. The monitor applies immediately.

### The default machine

`gaming` is an enhanced Apple //e set up for playing, the model that runs most
of the Apple II catalogue:

- 128Kb of RAM, expanded to 8Mb with a RAMWorks card, which the big game
  collections use
- A Mockingboard sound card in slot 4, for the games that play their music
  through it
- A FASTChip accelerator in slot 3, which the game launchers use to load faster
- A Disk II controller in slot 6, with DOS 3.3 in the first drive
- No mouse, and no VidHD card, since the core does not render super hi-res

Loading a hard disk image adds a SmartPort card in slot 7 by itself.

The other models are the regular izapple2 ones, see
[command_line.md](command_line.md) for what each of them is. They are all
usable, but the video cards the core cannot render are taken out of them too.

## Controls

### Keyboard

The core reads your keyboard directly, so just type. Most keys do the obvious
thing.

| Key | Apple II |
| --- | --- |
| Return | Return |
| Left arrow, Backspace | Left arrow |
| Right, Up, Down arrows | The matching arrows |
| Escape | Escape |
| Tab | Tab |
| Delete | Delete |
| Ctrl + letter | The control code, for example Ctrl-C to break |
| Left alt or option | Open apple |
| Right alt or option | Closed apple |

Reset is the frontend's own Reset, in *Quick Menu > Restart*, or whatever you
bound the reset hotkey to.

RetroArch keeps a good number of keys for its own hotkeys, and they never reach
the machine: `F1` opens the menu and `p` pauses the emulator instead of typing a
P. **Game Focus** is the switch for that: with it on, every key goes to the
Apple II and only the key that toggles it is left to the frontend.

Set it up in this order, the second step is easy to regret on its own:

1. *Settings > Input > Hotkeys > Game Focus (Toggle)* and bind it to a key you
   have. It comes bound to Scroll Lock, which no Mac keyboard has, so on a Mac
   you can turn Game Focus on and have nothing left that turns it off. Avoid the
   alt and option keys, the core uses those for the open and closed apple.
2. *Settings > Input > Auto Enable 'Game Focus' Mode* and set it to `Detect`.
   The core tells the frontend that it has a keyboard, so Game Focus comes on by
   itself while the emulator runs and stays off for the game cores.

If you are stuck in Game Focus with no working toggle, quit the frontend the way
the desktop does it, ⌘Q or the menu bar on macOS, which Game Focus does not
swallow.

### Gamepad

The first controller port drives the Apple II paddles:

| Control | Apple II |
| --- | --- |
| Left analog stick | Paddles 0 and 1 |
| D-pad | Paddles 0 and 1, for controllers with no stick |
| A | Button 0, same as the open apple |
| B | Button 1, same as the closed apple |
| X | Button 2 |

Only one joystick is emulated. The second pair of paddles always reads
centered.

## What is not supported

The core is deliberately smaller than the other frontends. If you need any of
this, use `izapple2sdl`.

- **Super hi-res graphics.** The VidHD card is removed from the machine, so the
  box art of the [Total Replay](https://archive.org/details/TotalReplay)
  collection does not show. Total Replay itself runs, it just looks like it does
  on a machine without a VidHD.
- **Videx 80 column video**, so the `ultraterm` model is not offered and the
  Videx cards are removed from the models that carry them.
- **The RGB card** and its extra graphic modes.
- **Savestates**, and therefore no rewind, no run ahead and no netplay. The
  emulator has no way to save the state of the machine, so this is not something
  the core can turn on.
- **The mouse.** The models with a mouse card run, but the mouse never moves.
- **Cheats**, which do not apply to a computer.

Everything else that izapple2 emulates works: the disk drives, the sound
including the Mockingboard, the language card, the 80 column card, the
accelerators and the rest.

## Video

The core hands the frontend a 564x192 image with a 4:3 aspect ratio. The
picture is not scaled, and the gaps between the scan lines are not drawn, so
that you can use whatever shader you prefer. A CRT shader such as `crt-geom` or
`crt-royale` gives a good period look on top of either monitor.

The choice of monitor is a core option and not a shader, because the two are
not the same thing. A green monitor does not just tint the picture: the Apple II
draws half width pixels on it that a colour television cannot show, so the
emulator has to render the screen differently from the start.

If the picture looks stretched, check that *Settings > Video > Scaling > Aspect
Ratio* is set to `Core Provided`.

## Troubleshooting

**The core does not appear in the list.** Use *Load Core > Install or Restore a
Core* and point at the library file. If it loads but shows up unnamed, or the
content dialog does not filter Apple II disks, the `izapple2_libretro.info`
file is not in the info directory.

**The core fails to load.** Almost always an architecture mismatch, a core
built for a different machine than the one running it. Rebuild on the target,
or cross compile with the right `CC`, `GOOS` and `GOARCH`.

**The frontend dies as soon as the core loads**, with a `SIGSEGV` in
`runtime.(*mheap).allocNeedsZero`. The frontend is being translated, Rosetta on
a Mac, and the Go runtime of the core cannot allocate in a process like that. A
core for both architectures does not help, the frontend itself has to run
natively. See [libretro_macos.md](libretro_macos.md).

**Some keys do nothing, or `p` pauses the emulator.** Those are hotkeys of the
frontend, turn on Game Focus, see [Keyboard](#keyboard).

**No sound during loading.** Expected while the disk is being read at speed,
see [Speed](#speed) above.

**Rewind and run ahead are greyed out.** The core has no savestates, see above.

**A game asks for disk 2, and the Disk Control menu is empty.** You loaded a
single disk image rather than an `.m3u` playlist listing all of them. Write the
playlist and load that instead, see
[Games on several diskettes](#games-on-several-diskettes).

**My saved game disappeared.** The saves live in the `.ovl` file next to the
other saves, see [Saved games](#saved-games). If you deleted it, or the frontend
is pointing at a different save directory than before, the disk loads as it came
and the saves are not there. Renaming the disk image also loses them, the
overlay is found by the name of the file.

**"the overlay belongs to another disk image".** The `.ovl` file was made from
a different image with the same name, so applying it would corrupt the disk and
the core refuses. Delete the `.ovl` file, losing what was saved in it, or put
back the image it was made from.

**A game cannot save.** It is probably on a WOZ or NIB diskette, which are read
only. Look for the same game as a DSK or PO.
