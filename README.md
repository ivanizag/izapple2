# izapple2 - Apple ][+, //e emulator

Portable emulator of an Apple II+ or //e. Written in Go.

## Features

- Models:
  - Apple ][+ with 48Kb of base RAM
  - Apple //e with 128Kb of RAM
  - Apple //e enhanced with 128Kb of RAM
  - Base64A clone with 48Kb of base RAM and paged ROM
- Storage
  - 16 Sector 5 1/4 diskettes. Uncompressed or compressed witth gzip or zip. Supported formats:
    - NIB (read only)
    - DSK
    - PO
    - [WOZ 1.0 or 2.0](storage/WozSupportStatus.md) (read only)
  - 3.5 disks in PO or 2MG format
  - Hard disk in HDV or 2MG format with ProDOS and SmartPort support
- Emulated extension cards:
  - DiskII controller (state machine based for WOZ files)
  - 16Kb Language Card
  - 256Kb Saturn RAM
  - 1Mb Memory Expansion Card (slinky)
  - RAMWorks style expansion Card (up to 16MB additional) (Apple //e only)
  - ThunderClock Plus real time clock
  - Apple //e 80 columns with 64Kb extra RAM and optional RGB modes
  - No Slot Clock based on the DS1216
  - Videx Videoterm 80 column card with the Videx Soft Video Switch (Apple ][+ only)
- Useful cards not emulating a real card
  - Bootable Smartport / ProDOS card
  - VidHd, limited to the ROM signature and SHR as used by Total Replay, only for //e models with 128Kb
  - FASTChip, limited to what Total Replay needs to set and clear fast mode
  - Mouse Card, emulates the entry points, not the softswitches.
  - Host console card. Maps the host STDIN and STDOUT to PR# and IN#

- Graphic modes:
  - Text 40 columns
  - Text 80 columns (Apple //e and Videx VideoTerm)
  - Low-Resolution graphics
  - Double-Width Low-Resolution graphics (Apple //e only)
  - High-Resolution graphics
  - Double-Width High-Resolution graphics (Apple //e only)
  - Super High Resolution (VidHD only)
  - Mixed mode
  - RGB card text 40 columns with 16 colors for foreground and background (mixable)
  - RGB card mode 11, mono 560x192
  - RGB card mode 12, ntsc 160*192
  - RGB card mode 13, ntsc 140*192 (regular DHGR)
  - RGB card mode 14, mix of modes 11 and 13 on the fly
- Displays:
  - Green monochrome monitor with half width pixel support
  - NTSC Color TV (extracting the phase from the mono signal)
  - RGB for Super High Resolution and RGB card
  - ANSI Console, avoiding the SDL2 dependency
  - Debug mode: shows four panels with actual screen, page1, page2 and extra info dependant of the video mode
- Tracing capabilities:
  - CPU execution disassembled
  - Softswitch reads and writes
  - ProDOS MLI calls
  - Apple Pascal BIOS calls
  - Smartport commands
  - BBC MOS calls when using [Applecorn](https://github.com/bobbimanners/)
- Other features:
  - Sound
  - Joystick support. Up to two joysticks or four paddles
  - Mouse support. No mouse capture needed
  - Adjustable speed
  - Fast disk mode to set max speed while using the disks
  - Single file executable with embedded ROMs and DOS 3.3
  - Pause (thanks a2geek)
  - Passes the [A2AUDIT 1.06](https://github.com/zellyn/a2audit) tests as II+, //e, and //e Enhanced.

By default the following configuration is launched:

- Enhanced Apple //e with 65c02 processor
- RAMWorks card with 80 column, RGB (with Video7 modes) and 8Gb RAM in aux slot
- Memory Expansion card with 1Gb in slot 1
- VidHD card (SHR support) in slot 2
- FASTChip Accelerator card in slot 3
- Mouse card in slot 4
- SmartPort card with 1 device in slot 5 (if an image is provided with -disk35)
- DiskII controller card in slot 6
- SmartPort card with 1 device in slot 7 (if an image is provided with -hd)

## Running the emulator

No installation required. [Download](https://github.com/ivanizag/izapple2/releases) the single file executable `izapple2xxx_xxx` for linux or Mac, SDL2 graphics or console. Build from source to get the latest features.

### Default mode

Execute without parameters to have an emulated Apple //e Enhanced with 128kb booting DOS 3.3 ready to run Applesoft:

``` terminal
casa@servidor:~$ ./izapple2sdl
```

![DOS 3.3 started](doc/dos33.png)

### Play games

Download a DSK or WOZ file or use an URL ([Asimov](https://www.apple.asimov.net/images/) is an excellent source):

``` terminal
casa@servidor:~$ ./izapple2sdl "https://www.apple.asimov.net/images/games/action/karateka/karateka (includes intro).dsk"
```

![Karateka](doc/karateka.png)

### Play the Total Replay collection

Download the excellent [Total Replay](https://archive.org/details/TotalReplay) compilation by
[a2-4am](https://github.com/a2-4am/4cade):

``` terminal
casa@servidor:~$ ./izapple2sdl Total\ Replay\ v4.0.hdv
```

Displays super hi-res box art as seen with the VidHD card.

![Total Replay](doc/totalreplay.png)

### Terminal mode

To run text mode right on the terminal without the SDL2 dependency, use `izapple2console`. It runs on the console using ANSI escape codes. Input is sent to the emulated Apple II one line at a time:

``` terminal
casa@servidor:~$ ./izapple2console -model 2plus

############################################
#                                          #
#                APPLE II                  #
#                                          #
#     DOS VERSION 3.3  SYSTEM MASTER       #
#                                          #
#                                          #
#            JANUARY 1, 1983               #
#                                          #
#                                          #
# COPYRIGHT APPLE COMPUTER,INC. 1980,1982  #
#                                          #
#                                          #
# ]10 PRINT "HELLO WORLD"                  #
#                                          #
# ]LIST                                    #
#                                          #
# 10  PRINT "HELLO WORLD"                  #
#                                          #
# ]RUN                                     #
# HELLO WORLD                              #
#                                          #
# ]_                                       #
#                                          #
#                                          #
############################################
Line:

```

### Keys

- Ctrl-F1: Reset button
- F5: Toggle speed between real and fastest
- Ctrl-F5: Show current speed in Mhz
- F6: Toggle between NTSC color TV and green phosphor monochrome monitor
- F7: Show the video mode and a split screen with the views for NTSC color TV, page 1, page 2 and extra info.
- F10: Cycle character generator code pages. Only if the character generator ROM has more than one 2Kb page.
- Ctrl-F10: Show the charater map for the current character generator page.
- Shift-F10: When showing the character map, use altText.
- F11: Toggle on and off the trace to console of the CPU execution
- F12: Save a screen snapshot to a file `snapshot.png`
- Pause: Pause the emulation

Only valid on SDL mode

### Command line options

```terminal
  -charRom string
        rom file for the character generator (default "<default>")
  -consoleCardSlot int
    	  slot for the host console card. -1 for none (default -1)
  -disk string
        file to load on the first disk drive (default "<internal>/dos33.dsk")
  -disk2Slot int
        slot for the disk driver. -1 for none. (default 6)
  -disk35 string
        file to load on the SmartPort disk (slot 5)
  -diskRom string
        rom file for the disk drive controller (default "<internal>/DISK2.rom")
  -diskb string
        file to load on the second disk drive
  -fastChipSlot int
        slot for the FASTChip accelerator card, -1 for none (default 3)
  -forceCaps
        force all letters to be uppercased (no need for caps lock!)
  -fastDisk
        set fast mode when the disks are spinning (default true)
  -hd string
        file to load on the boot hard disk
  -hdSlot int
        slot for the hard drive if present. -1 for none. (default -1)
  -languageCardSlot int
        slot for the 16kb language card. -1 for none
  -memoryExpSlot int
        slot for the Memory Expansion card with 1GB. -1 for none (default 1)
  -mhz float
        cpu speed in Mhz, use 0 for full speed. Use F5 to toggle. (default 1.0227142857142857)
  -model string
        set base model. Models available 2plus, 2e, 2enh, base64a (default "2enh")
  -mouseCardSlot int
    	  slot for the Mouse card. -1 for none (default 4)
  -nsc int
        add a DS1216 No-Slot-Clock on the main ROM (use 0) or a slot ROM. -1 for none (default -1)
  -panicSS
        panic if a not implemented softswitch is used
  -profile
        generate profile trace to analyse with pprof
  -ramworks int
        memory to use with RAMWorks card, 0 for no card, max is 16384 (default 8192)
  -rgb
        emulate the RGB modes of the 80col RGB card for DHGR (default true)
  -rom string
        main rom file (default "<default>")
  -saturnCardSlot int
        slot for the 256kb Saturn card. -1 for none (default -1)
  -sequencer
        use the sequencer based Disk II card
  -thunderClockCardSlot int
        slot for the ThunderClock Plus card. -1 for none (default 4)
  -traceBBC
        trace BBC MOS API calls used with Applecorn, skip console I/O calls
  -traceBBCFull
        trace BBC MOS API calls used with Applecorn
  -traceCpu
        dump to the console the CPU execution. Use F11 to toggle.
  -traceHD
        dump to the console the hd/smartport commands
  -traceMLI
        dump to the console the calls to ProDOS machine language interface calls to $BF00
  -tracePascal
        dump to the console the calls to the Apple Pascal BIOS
  -traceSS
        dump to the console the sofswitches calls
  -traceSSReg
        dump to the console the sofswitch registrations
  -vidHDSlot int
        slot for the VidHD card, only for //e models. -1 for none (default 2)
  -videxCardSlot int
    	  slot for the Videx Videoterm 80 columns card. For pre-2e models. -1 for none (default 3)


```

## Building from source

### Linux

Besides having a working Go installation, install the SDL2 developer files. Run:

``` terminal
git clone github.com/ivanizag/izapple2
cd izapple2/frontend/a2sdl
go build .
```

### Windows 

On Windows, CGO needs a gcc compiler. Install [mingw-w64](http://mingw-w64.org/doku.php/download/mingw-builds) and the [SDL2 developer files](https://www.libsdl.org/release/) for mingw-64.

Run:

``` terminal
git clone github.com/ivanizag/izapple2
cd izapple2\frontend\a2sdl
go build .
```

To run in Windows, copy the file `SDL2.dll` on the same folder as `a2sdl.exe`. The latest `SDL2.dll` can be found in the [Runtime binary for Windows 64-bit](https://www.libsdl.org/download-2.0.php).

### Use docker to cross compile for Linux and Windows

To create executables for Linux and Windows without installing Go, SDL2 or the Windows cross compilation toosl, run:

``` terminal
cd docker
./build.sh
```
