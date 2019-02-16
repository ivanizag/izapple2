package main

import "go6502/apple2"

func main() {
	romFile := "../roms/apple.rom"
	//romFile := "../roms/APPLE2.ROM"

	log := false

	apple2.Run(romFile, log)
}
