# Apple II+ emulator

Portable emulator of an Apple II+. Written in Go.

## Features

- Apple II+ with 48Kb of base RAM
- Sound
- 16 Sector diskettes in DSK format
- Emulated extension cards:
    - DiskII controller
    - 16Kb Language Card
    - 256kb Saturn RAM Card
- Graphic modes:
    - Text, Lores and Hires
- Displays:
    - Green monochrome monitor with half width pixel support
    - NTSC Color TV (extracting the phase from the mono signal)
    - ANSI Console, avoiding the SDL2 dependency
- Adjustable speed.
- Fast disk mode to set max speed while using the disks. 
- Single file executable with embedded ROMs and DOS 3.3


## Running the emulator

No installation required. [Download](https://github.com/ivanizag/apple2/releases) the single file executable `apple2xxx_xxx` for linux or Mac, SDL2 graphics or console.

### Default mode

Execute without parameters to have an emulated Apple II+ with 64kb booting DOS 3.3 ready to run Applesoft:
```
casa@servidor:~$ ./apple2sdl
```

![DOS 3.3 started](doc/dos33.png)

### Play games
Download an DSK file ([Asimov](https://mirrors.apple2.org.za/ftp.apple.asimov.net/images/) is an excellent source) and use the `-disk` parameter.
```
casa@servidor:~$ ./apple2sdl -disk ~/Downloads/karateka.dsk
```
![Karateka](doc/karateka.png)


### Terminal mode
To run text mode right on the terminal without the SDL2 dependency just run `apple2console`. It runs on the console using ANSI escape codes. Input is sent to the emulated Apple II one line at a time: 
```
casa@servidor:~$ ./apple2console

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

- F5: Toggle speed between real and fastest
- F6: Toggle between NTSC color TV and green phosphor monochrome monitor
- F7: Save current state to disk
- F8: Restore state from disk
- F12: Save a screen snapshot to a file `snapshot.png`

Only valid on SDL mode

### Command line options

```
  -charRom string
    	rom file for the disk drive controller (default "<internal>/Apple2rev7CharGen.rom")
  -disk string
    	file to load on the first disk drive (default "<internal>/dos33.dsk")
  -disk2Slot int
    	slot for the disk driver. -1 for none. (default 6)
  -diskRom string
    	rom file for the disk drive controller (default "<internal>/DISK2.rom")
  -dumpChars
    	shows the character map
  -fastDisk
    	set fast mode when the disks are spinning (default true)
  -languageCardSlot int
    	slot for the 16kb language card. -1 for none
  -mhz float
    	cpu speed in Mhz, use 0 for full speed. Use F5 to toggle. (default 1.0227142857142857)
  -mono
    	emulate a green phosphor monitor instead of a NTSC color TV. Use F6 to toggle.
  -panicss
    	panic if a not implemented softswitch is used
  -rom string
    	main rom file (default "<internal>/Apple2_Plus.rom")
  -saturnCardSlot int
    	slot for the 256kb Saturn card. -1 for none (default -1)
```

## Building from source

### apple2console

The only dependency is having a working Go installation.
Run:
```
$ go get github.com/ivanizag/apple2/apple2console 
$ go build github.com/ivanizag/apple2/apple2console 
``` 

### apple2sdl

Besides having a working Go installation, install the SDL2 developer files.

Run:
```
$ go get github.com/ivanizag/apple2/apple2sdl
$ go build github.com/ivanizag/apple2/apple2sdl 
