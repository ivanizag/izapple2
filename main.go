package main

import (
	"go6502/apple2"
	"go6502/apple2sdl"
)

func main() {
	//romFile := "apple2/romdumps/Apple2.rom"
	romFile := "apple2/romdumps/Apple2_Plus.rom"
	//romFile := "apple2/romdumps/Apple2e.rom"
	disk2RomFile := "apple2/romdumps/DISK2.rom"
	diskImage := "../dos33.dsk"

	log := false
	sdl := true
	a := apple2.NewApple2(romFile)
	a.AddDisk2(disk2RomFile, diskImage)
	if sdl {
		apple2sdl.SDLRun(a)
	} else {
		a.Run(log, true)
	}
}
