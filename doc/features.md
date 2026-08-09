# Features

- Models:
  - Apple ][+ with 48Kb of base RAM
  - Apple //e with 128Kb of RAM
  - Apple //e enhanced with 128Kb of RAM
  - Base64A clone with 48Kb of base RAM and paged ROM
  - Basis 108 clone (partial)
- Storage
  - 16 Sector 5 1/4 diskettes. Uncompressed or compressed with gzip or zip. Supported formats:
    - NIB (read only)
    - DSK
    - PO
    - [WOZ 1.0 or 2.0](../storage/WozSupportStatus.md) (read only)
  - 13 Sector 5 1/4 diskettes. Uncompressed or compressed with gzip or zip. Supported formats:
    - NIB (read only)
    - [WOZ 2.0](../storage/WozSupportStatus.md) (read only)
  - 3.5 disks in PO or 2MG format
  - Hard disk in HDV or 2MG format with ProDOS and SmartPort support
  - Cassette tape input from WAV recordings

- Emulated extension cards:
  - DiskII controller (state machine based for WOZ files)
  - 16Kb Language Card
  - 256Kb Saturn RAM
  - Parallel Printer Interface card
  - 1Mb Memory Expansion Card (slinky)
  - RAMWorks style expansion Card (up to 16MB additional) (Apple //e only)
  - ThunderClock Plus real time clock
  - Apple //e 80 columns card with 64Kb extra RAM and optional RGB modes
  - No Slot Clock based on the DS1216
  - Videx Videoterm 80 column card with the Videx Soft Video Switch (Apple ][+ only)
  - Videx Ultraterm 80 to 160 column card with integrated Video Switch
  - SwyftCard (Apple //e only)
  - Brain Board
  - Brain Board II
  - MultiROM card
  - Dan ][ Controller card
  - ProDOS ROM card
  - Microsoft Z80 Softcard using the [Z80](https://github.com/koron-go/z80) emulation from Koron
  - Mockingboard A sound card
- Useful cards not emulating a real card
  - Bootable SmartPort / ProDOS card with the following smartport devices:
      - Block device (hard disks)
      - Fujinet network device (supports only http(s) with GET and JSON)
      - Fujinet clock (not in Fujinet upstream)
  - VidHD, limited to the ROM signature and SHR as used by Total Replay, only for //e models with 128Kb
  - FASTChip, limited to what Total Replay needs to set and clear fast mode
  - Mouse Card, emulates the entry points, not the softswitches.
  - Host console card. Maps the host STDIN and STDOUT to PR# and IN#
  - ROMXe, limited to font switching

- Graphic modes:
  - Text 40 columns
  - Text 80 columns Apple //e
  - Text 80 columns Videx VideoTerm
  - Text up to 160 columns and 48 lines Videx UltraTerm
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
  - Debug mode: shows four panels with actual screen, page1, page2 and extra info dependent on the video mode
- Tracing capabilities:
  - CPU execution disassembled
  - Softswitch reads and writes
  - ProDOS MLI calls
  - Apple Pascal BIOS calls
  - SmartPort commands
  - BBC MOS calls when using [Applecorn](https://github.com/bobbimanners/)
- Other features:
  - Sound
  - Joystick support. Up to two joysticks or four paddles
  - Mouse support. No mouse capture needed
  - Adjustable speed
  - Fast disk mode to set max speed while using the disks
  - Save directory with `-saveDir`: what the software writes goes to an overlay file per disk, keeping the images untouched and making writable the disks loaded from a compressed file, an URL or the embedded resources
  - Single file executable with embedded ROMs and DOS 3.3
  - Pause (thanks a2geek)
  - Passes the [A2AUDIT 1.06](https://github.com/zellyn/a2audit) tests as II+, //e, and //e Enhanced.
  - Partial pass of the [ProcessorTests](https://github.com/TomHarte/ProcessorTests) for 6502 and 65c02. Failing test 6502/v1/20_55_13; flags N and V issues with ADC; and missing some undocumented 6502 opcodes.

## The default configuration

By default the following configuration is launched:

- Enhanced Apple //e with 65c02 processor
- RAMWorks card with 80 columns and 8Mb of RAM in the aux slot
- No Slot Clock on the main ROM
- Language card in slot 0
- VidHD card (SHR support) in slot 2
- FASTChip accelerator card in slot 3
- Mockingboard sound card in slot 4
- DiskII controller card with DOS 3.3 in slot 6

Slots 1, 5 and 7 are empty, and the RGB modes of the 80 columns card are added
with `-rgb`. Diskettes and hard disks named in the command line take the slots
they need: diskettes go to slot 6 and then slot 5, block devices to a SmartPort
card in slot 7 and then slot 5.

Run `izapple2 -showConfig` to see the configuration a command line ends up
with, without starting the machine.

Everything here is chosen with the [command line options](command_line.md), and
what each frontend can do with it is in its own page, listed in the
[README](../README.md#frontends).
