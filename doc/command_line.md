# Command Line Configuration Guide

This guide explains how to configure the izapple2 emulator using command line options.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Models](#models)
- [Slot Configuration](#slot-configuration)
- [Simplified Filename Configuration](#simplified-filename-configuration)
- [Reference](#reference)
- [Examples](#examples)
- [Tips](#tips)

## Basic Usage

```bash
izapple2 [options] [files...]
```

The emulator can be started with various command line options to customize the hardware configuration, and optionally with disk image files to load.

## Models

Models are pre-configured hardware setups that define the Apple II variant to emulate. Each model specifies the board type, CPU, ROM, default cards, and other hardware characteristics. It also load a default set of diskettes.

### Available Models

Use the `-model` flag to select a model:

```bash
izapple2 -model=2enh
```

**Available models:**

- `2` - Apple ][
- `2e` - Apple IIe
- `2enh` - Apple //e (default, enhanced version)
- `2plus` - Apple ][+
- `base64a` - Base 64A
- `basis108` - Basis 108
- `cpm` - Apple ][+ with CP/M
- `cpm3` - Apple //e with CP/M 3.0
- `cpm65` - Apple //e with CPM-65
- `desktop` - Apple II DeskTop
- `dos32` - Apple ][ with 13 sectors disk adapter and DOS 3.2x
- `pascal` - Apple //e with Apple Pascal 1.3
- `prodos` - Apple //e ProDOS
- `swyft` - Swyft
- `ultraterm` - Apple ][+ with Videx Ultraterm demo

For a complete list of available models, see the `configs/` directory or run:

```bash
izapple2 -h
```

### Custom Models

You can create custom model configuration files (`.cfg` format) and load them by specifying the filename:

```bash
izapple2 -model=myconfig.cfg
```

Set the `IZAPPLE2_CUSTOM_MODEL` environment variable to specify a default location for custom model files.

## Slot Configuration

The Apple II has 8 expansion slots (0-7) where peripheral cards can be installed. Each slot can be configured independently.

### Full Slot Configuration Syntax

The full syntax for configuring a slot is:

```bash
-s<slot> <card_name>[,<param1>=<value1>,<param2>=<value2>,...]
```

**Example:**
```bash
izapple2 -s6 diskii,disk1="mydisk.dsk",disk2="backup.dsk"
```

### Available Cards

Common cards include:

- `diskii` - Disk II floppy drive controller
- `smartport` - SmartPort hard disk controller
- `mouse` - Mouse card
- `parallel` - Parallel printer card
- `vidhd` - VidHD graphics card
- `fastchip` - Accelerator card
- `language` - Language card (16KB RAM expansion)
- `empty` - No card installed

For a complete list of available cards and their parameters, see the individual card documentation files in the `doc/` directory or run:

```bash
izapple2 -h
```

### Slot Defaults

The default model (`2enh`) configures slots as follows:

- **Slot 0:** Language card
- **Slot 1:** Empty
- **Slot 2:** VidHD
- **Slot 3:** FastChip
- **Slot 4:** Empty
- **Slot 5:** Empty
- **Slot 6:** Disk II with DOS 3.3
- **Slot 7:** Empty

## Simplified Filename Configuration

### Positional Filename Arguments

The simplest way to load disk images is to pass them as positional arguments (without specifying a slot):

```bash
izapple2 disk1.dsk disk2.dsk
```

The emulator will automatically configure slots based on file types:
- **Diskettes:** Configured in slots 6 and 5 (up to 4 diskettes total)
- **Block devices:** Configured in slots 7 and 5 (up to 8 devices total)

#### How It Works

The emulator automatically:
1. Detects whether each file is a diskette (`.dsk`, `.do`, `.po` in diskette format) or block device
2. Creates the appropriate card configuration:
   - **Diskettes** → `diskii` card in slot 6 (first two) and slot 5 (next two)
   - **Block devices** → `smartport` card in slot 7 (first) and slot 5 (remaining)

#### Examples

```bash
# Single diskette → slot 6
izapple2 mydisk.dsk

# Two diskettes → slot 6 (both disks)
izapple2 disk1.dsk disk2.dsk

# Three diskettes → slot 6 (first two), slot 5 (third)
izapple2 disk1.dsk disk2.dsk disk3.dsk

# Four diskettes → slot 6 (first two), slot 5 (last two)
izapple2 disk1.dsk disk2.dsk disk3.dsk disk4.dsk
```

### Partial Slot Configuration

If the default slot assignment doesn't meet your needs, you can configure specific slots using just filenames. The emulator will automatically detect the file type and create the appropriate card configuration.

#### Syntax

Instead of the full configuration:
```bash
izapple2 -s4 diskii,disk1="disk.dsk"
```

You can simply use:
```bash
izapple2 -s4 disk.dsk
```

#### Multiple Files in One Slot

You can specify multiple files separated by commas:

```bash
# Two diskettes in slot 6
izapple2 -s6 disk1.dsk,disk2.dsk

# Equivalent to:
izapple2 -s6 diskii,disk1="disk1.dsk",disk2="disk2.dsk"
```

**Limitations:**
- Disk II cards support a maximum of 2 diskettes per slot
- You cannot mix diskettes and block devices in the same slot

### Disk Aliases

The emulator provides convenient aliases for common disk images:

- `dos33` → `<internal>/dos33.dsk`
- `prodos` → `<internal>/ProDOS_2_4_3.po`
- `cpm` → `<internal>/cpm_2.20B_56K.po`
- `cardcat` → `<internal>/Card Cat 1.7.dsk`

**Example:**
```bash
izapple2 -s6 dos33
# Equivalent to: izapple2 -s6 diskii,disk1="<internal>/dos33.dsk"
```

## Reference

Complete command line options reference:

<!-- doc/usage.txt start -->
```terminal
Usage:  izapple2 [file]
  file
    	path to image to use on the boot device
  -charrom string
    	rom file for the character generator (default "<internal>/Apple IIe Video Enhanced.bin")
  -cpu string
    	cpu type, can be '6502' or '65c02' (default "65c02")
  -forceCaps
    	force all letters to be uppercased (no need for caps lock!)
  -model string
    	set base model (default "2enh")
  -mods string
    	comma separated list of mods applied to the board, available mods are 'shift', 'four-colors
  -nsc string
    	add a DS1216 No-Slot-Clock on the main ROM (use 'main') or a slot ROM (default "main")
  -profile
    	generate profile trace to analyse with pprof
  -ramworks string
    	memory to use with RAMWorks card, max is 16384 (default "8192")
  -rgb
    	emulate the RGB modes of the 80col RGB card for DHGR
  -rom string
    	main rom file (default "<internal>/Apple2e_Enhanced.rom")
  -romx
    	emulate a RomX
  -s0 string
    	slot 0 configuration. (default "language")
  -s1 string
    	slot 1 configuration. (default "empty")
  -s2 string
    	slot 2 configuration. (default "vidhd")
  -s3 string
    	slot 3 configuration. (default "fastchip")
  -s4 string
    	slot 4 configuration. (default "empty")
  -s5 string
    	slot 5 configuration. (default "empty")
  -s6 string
    	slot 6 configuration. (default "diskii,disk1=<internal>/dos33.dsk")
  -s7 string
    	slot 7 configuration. (default "empty")
  -showConfig
    	show the calculated configuration and exit
  -speed string
    	cpu speed in Mhz, can be 'ntsc', 'pal', 'full' or a decimal nunmber (default "ntsc")
  -trace string
    	trace CPU execution with one or more comma separated tracers (default "none")

The available pre-configured models are:
  2: Apple ][
  2e: Apple IIe
  2enh: Apple //e
  2plus: Apple ][+
  base64a: Base 64A
  basis108: Basis 108
  cpm: Apple ][+ with CP/M
  cpm3: Apple //e with CP/M 3.0
  cpm65: Apple //e with CPM-65
  desktop: Apple II DeskTop
  dos32: Apple ][ with 13 sectors disk adapter and DOS 3.2x
  pascal: Apple //e with Apple Pascal 1.3
  prodos: Apple //e Prodos
  swyft: swyft
  ultraterm: Apple ][+ with Videx Ultraterm demo
Custom models may be specified by filename.  Use 'IZAPPLE2_CUSTOM_MODEL' to set default location.

The available cards are:
  brainboard: Firmware card. It has two ROM banks
  brainboard2: Firmware card. It has up to four ROM banks
  dan2sd: Apple II Peripheral Card that Interfaces to a ATMEGA328P for SD card storage
  diskii: Disk II interface card
  diskiiseq: Disk II interface card emulating the Woz state machine
  fastchip: Accelerator card for Apple IIe (limited support)
  fujinet: SmartPort interface card hosting the Fujinet
  language: Language card with 16 extra KB for the Apple ][ and ][+
  memexp: Memory expansion card
  mouse: Mouse card implementation, does not emulate a real card, only the firmware behaviour
  multirom: Multiple Image ROM card
  parallel: Card to dump to a file what would be printed to a parallel printer
  prodosblock: ProDOS block device interface card
  prodosromcard3: A bootable 4 MB ROM card by Ralle Palaveev
  prodosromdrive: A bootable 1 MB solid state disk by Terence Boldt
  saturn: RAM card with 128Kb, it's like 8 language cards
  smartport: SmartPort interface card
  softswitchlogger: Card to log softswitch accesses
  swyftcard: Card with the ROM needed to run the Swyftcard word processing system
  thunderclock: Clock card
  videx: Videx Videoterm compatible 80 columns card
  videxultraterm: Videx Utraterm compatible 80 columns card
  vidhd: Firmware signature of the VidHD card to trick Total Replay to use the SHR mode
  z80softcard: Microsoft Z80 SoftCard to run CP/M

The available tracers are:
  cpm: Trace CPM BDOS calls
  cpm65: Trace CPM65 BDOS calls skipping terminal IO
  cpm65full: Trace CPM65 BDOS calls
  cpu: Trace CPU execution
  mli: Trace ProDOS MLI calls
  mos: Trace MOS calls with Applecorn skipping terminal IO
  mosfull: Trace MOS calls with Applecorn
  panicss: Panic on unimplemented softswitches
  rom: Trace monitor ROM calls
  ss: Trace sotfswiches calls
  ssreg: Trace sotfswiches registrations
  ucsd: Trace UCSD system calls

```
<!-- doc/usage.txt end -->


## Examples

### Basic Examples

```bash
# Start with default configuration
izapple2

# Start with a specific disk
izapple2 mydisk.dsk

# Use a different model
izapple2 -model=2plus

# Show current configuration
izapple2 -showConfig
```

### Slot Configuration Examples

```bash
# Configure slot 4 with a diskette (simplified)
izapple2 -s4 games.dsk

# Configure slot 4 with two diskettes (simplified)
izapple2 -s4 disk1.dsk,disk2.dsk

# Configure slot 7 with a hard disk (full syntax)
izapple2 -s7 smartport,image1="harddisk.po"

# Configure multiple slots
izapple2 -s4 disk1.dsk -s5 disk2.dsk -s7 smartport,image1="hd.po"

# Mix simplified and full syntax
izapple2 -s4 mydisk.dsk -s6 diskii,disk1="boot.dsk",disk2="data.dsk"
```

### Advanced Examples

```bash
# Apple IIe with custom speed and tracing
izapple2 -model=2enh -speed=2.0 -trace=cpu,ss

# Maximum RAM configuration
izapple2 -ramworks=16384

# Custom ROM with RGB output
izapple2 -rom=custom.rom -rgb

# Multiple disks with different configurations
izapple2 -s6 dos33 -s5 disk1.dsk,disk2.dsk -s7 smartport,image1="hd.po"
```

### Positional Arguments Examples

```bash
# Boot from a specific disk
izapple2 mydisk.dsk

# Load multiple diskettes
izapple2 boot.dsk data1.dsk data2.dsk

# Combine positional arguments with slot configuration
izapple2 boot.dsk -s7 smartport,image1="harddisk.po"
```

## Tips

1. **Flags must come before positional arguments** in the command line
2. **Use quotes** around filenames with spaces: `-s6 "my disk.dsk"`
3. **Boolean flags** use `true`, `false` or nothing: `-showConfig=true` is the same as `-showConfig` 
4. **View all options** with `izapple2 -h`
5. **Test configurations** with `-showConfig` to see the final setup without starting the emulator


