package main

import (
	"flag"
	"go6502/apple2"
	"go6502/apple2sdl"
)

func main() {
	romFile := flag.String(
		"rom",
		"apple2/romdumps/Apple2_Plus.rom",
		"main rom file")
	disk2RomFile := flag.String(
		"diskRom",
		"apple2/romdumps/DISK2.rom",
		"rom file for the disk drive controller")
	disk2Slot := flag.Int(
		"disk2Slot",
		6,
		"slot for the disk driver. 0 for none.")
	diskImage := flag.String(
		"disk",
		"../dos33.dsk",
		"file to load on the first disk drive")
	cpuClock := flag.Float64(
		"mhz",
		apple2.CpuClockMhz,
		"cpu speed in Mhz, use 0 for full speed. Use F5 to toggle.")
	charRomFile := flag.String(
		"charRom",
		"apple2/romdumps/Apple2rev7CharGen.rom",
		"rom file for the disk drive controller")
	useSdl := flag.Bool(
		"sdl",
		true,
		"use SDL")
	stdoutScreen := flag.Bool(
		"stdout",
		false,
		"show the text screen on the standard output")
	mono := flag.Bool(
		"mono",
		false,
		"emulate a green phosphor monitor instead of a NTSC color TV. Use F6 to toggle.",
	)
	fastDisk := flag.Bool(
		"fastDisk",
		true,
		"set fast mode when the disks are spinning",
	)

	panicSS := flag.Bool(
		"panicss",
		false,
		"panic if a not implemented softswitch is used")
	dumpChars := flag.Bool(
		"dumpChars",
		false,
		"shows the character map",
	)
	flag.Parse()

	if *dumpChars {
		cg := apple2.NewCharacterGenerator(*charRomFile)
		cg.Dump()
		return
	}

	log := false
	a := apple2.NewApple2(*romFile, *charRomFile, *cpuClock, !*mono, *fastDisk, *panicSS)
	//a.AddLanguageCard(0)
	a.AddSaturnCard(0)
	if *disk2Slot > 0 {
		a.AddDisk2(*disk2Slot, *disk2RomFile, *diskImage)
	}
	if *useSdl {
		a.ConfigureStdConsole(false, *stdoutScreen)
		apple2sdl.SDLRun(a)
	} else {
		a.ConfigureStdConsole(true, true)
		a.Run(log)
	}
}
