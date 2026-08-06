# Running the libretro core on a Mac

From nothing to the core running in RetroArch, with the checks worth doing on
the way. See [libretro.md](libretro.md) for what the core is and what it
supports.

## 1. Install RetroArch

**Use `retroarch-metal`, not `retroarch`.**

``` terminal
brew install --cask retroarch-metal
```

The two casks are the same version and the graphics API makes no difference to
this core, which hands over a finished picture and never touches it. What
matters is the architecture: `retroarch-metal` is a universal build, while
`retroarch` is Intel only.

That is not a matter of taste. On an Apple Silicon Mac the Intel build runs
under Rosetta, and the Go runtime inside the core cannot allocate memory in a
translated process of that size. It dies immediately with

``` terminal
SIGSEGV: segmentation violation
PC=0x1228a0111 m=1 sigcode=1 addr=0x0
runtime.(*mheap).allocNeedsZero(...)
```

taking RetroArch with it. A core built for both architectures does not help,
because the problem is the translation and not the slice that gets loaded.

Check what you have:

``` terminal
lipo -archs /Applications/RetroArch.app/Contents/MacOS/RetroArch
```

`x86_64 arm64` is what you want. Plain `x86_64` on an Apple Silicon Mac is the
Intel cask, and the core will not run in it.

## 2. Build the core

``` terminal
git clone https://github.com/ivanizag/izapple2
cd izapple2/frontend/a2libretro
make universal
```

`make universal` builds for both architectures and prints what it made. Plain
`make` builds for the host only, which is fine on an Intel Mac and fine for an
arm64 RetroArch, but the universal one covers either.

## 3. Install it

``` terminal
mkdir -p ~/Library/Application\ Support/RetroArch/cores
mkdir -p ~/Library/Application\ Support/RetroArch/info
cp izapple2_libretro.dylib ~/Library/Application\ Support/RetroArch/cores/
cp izapple2_libretro.info ~/Library/Application\ Support/RetroArch/info/
```

The `.info` file is what makes RetroArch show the core by name with its list of
what it can and cannot do. Without it the core still works, but it appears
unnamed and the content dialog will not filter Apple II disks.

## 4. Let the keyboard reach the Apple II

This one is not optional on a computer core, and it is worth doing before you
start typing.

RetroArch keeps a good number of keys for its own hotkeys, and they never reach
the machine. `F1` opens the menu and, the one that will puzzle you, **`p` pauses
the emulator instead of typing a P**. Its answer is *Game Focus*: with it on,
every key goes to the core and RetroArch keeps only the key that toggles it.

**Rebind the toggle first.** It comes bound to Scroll Lock, which no Mac
keyboard has, so out of the box you can turn Game Focus on and have no way to
turn it off again.

1. *Settings > Input > Hotkeys > Game Focus (Toggle)*, press it and then press
   `F12`, or anything else free. Do not use the alt or option keys, the core
   uses those for the open and closed apple.
2. *Settings > Input > Auto Enable 'Game Focus' Mode* and set it to `Detect`.
   The core tells the frontend that it has a keyboard, so Game Focus comes on by
   itself whenever the emulator is running, and stays off for your game cores.
   `On` also works and applies to every core, where F12 turns it off.

The menu search, the magnifier or `/`, finds both if you type `focus`.

If you do end up stuck in Game Focus with no working toggle, **⌘Q** or the macOS
menu bar quits RetroArch. Those are handled by macOS and not by RetroArch's
input, so Game Focus does not swallow them.

## 5. Check that it works

Each of these checks something different, so it is worth going through them in
order the first time.

1. **Load Core**. "Apple II (izapple2)" is in the list, which means the `.info`
   file was found. Pick it.
2. **Start Core**. The Apple //e boots DOS 3.3 and lands on the `]` prompt after
   a couple of seconds. The picture, the disk emulation and the frame timing all
   work.
3. Type `PRINT 2+2` and Return. It answers `4`. The keyboard is reaching the
   machine, so Game Focus is doing its job.
4. *Settings > Video > Scaling*, set **Aspect Ratio** to `Core Provided`. The
   picture is 4:3 and the text is not stretched.
5. *Quick Menu > Core Options*, set **Monitor** to `green`. The screen turns
   green phosphor at once, with no reset.
6. Type `CATALOG` and Return. The files of the DOS 3.3 diskette are listed, so
   the drive is being read. Press a key when the listing pauses to see the rest.
7. Type `SAVE HELLO` and Return, then look in the save directory, which
   *Settings > Directory > Saves* shows and is `~/Documents/RetroArch/saves` by
   default. There is a `dos33.dsk.ovl` of a few kilobytes: the save went to an
   overlay and the DOS 3.3 built into the core was not touched.
8. **Load Content** and pick a `.dsk` or `.woz` game. It boots.
9. For a game on several diskettes, load an `.m3u` listing them and open
   *Quick Menu > Disk Control*. The diskettes are there by name, and ejecting,
   selecting another one and inserting swaps it.

## When something goes wrong

**RetroArch dies as soon as the core loads**, with a `SIGSEGV` in
`runtime.(*mheap).allocNeedsZero` in the log. The Intel RetroArch under Rosetta,
see step 1. Install `retroarch-metal`.

**The core does not appear in Load Core.** Wrong directory, check
*Settings > Directory > Cores*. If it loads but shows up unnamed, the `.info`
file is not in the info directory.

**Some keys do nothing, or `p` pauses the emulator.** Game Focus is off, see
step 4.

**Nothing gets you out of Game Focus.** The toggle is still on Scroll Lock.
⌘Q to quit, then rebind it.

To see what the core itself has to say, run RetroArch from a terminal instead of
from the Finder, so that its output is not thrown away:

``` terminal
/Applications/RetroArch.app/Contents/MacOS/RetroArch 2>&1 | tee /tmp/retroarch.log
```
