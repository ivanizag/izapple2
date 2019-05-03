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
	diskImage := flag.String(
		"disk",
		"../dos33.dsk",
		"file to load on the first disk drive")
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

	//romFile := "apple2/romdumps/Apple2.rom"
	//romFile := "apple2/romdumps/Apple2_Plus.rom"
	//romFile := "apple2/romdumps/Apple2e.rom"
	//disk2RomFile := "apple2/romdumps/DISK2.rom"
	//diskImage := "../dos33.dsk"
	//diskImage := "../Apex II - Apple II Diagnostic (v4.7-1986).DSK"
	//diskImage := "../A2Diag.v4.1.SDK"

	if *dumpChars {
		cg := apple2.NewCharacterGenerator(*charRomFile)
		cg.Dump()
		return
	}

	log := false
	a := apple2.NewApple2(*romFile, *charRomFile, *panicSS)
	a.AddDisk2(*disk2RomFile, *diskImage)
	if *useSdl {
		a.ConfigureStdConsole(false, *stdoutScreen)
		apple2sdl.SDLRun(a)
	} else {
		a.ConfigureStdConsole(true, true)
		a.Run(log)
	}
}
