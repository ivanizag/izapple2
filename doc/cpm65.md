# CPM-65

CPM-65 is a native port of Digital Research's seminal 1977 operating system CP/M to the 6502. See https://github.com/davidgiven/cpm65

It runs on an Apple IIe with 80 columns.
To run use the preconfigured model `cpm65`.

Usefull commands:

- `ìzapple -model cpm65` : To run the preconfigured setup
- `izapple apple2e.po`: To use the disk apple2e.po to run any release from [cpm65 releases](https://github.com/davidgiven/cpm65/releases/tag/dev)
- `ìzapple -model cpm65 -trace cpm65` : To trace the BDOS calls, skipping the console related calls
- `ìzapple -model cpm65 -trace cpm65full` : To trace all the BDOS calls.

Todo:
- The address for the BIOS and BDOS entrypoints is hardwired. It should be extracted somehow from the running memory.
