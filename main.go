package main

import "go6502/apple2"

func main() {
	//romFile := "apple2/romdumps/Apple2.rom"
	romFile := "apple2/romdumps/Apple2_Plus.rom"
	//romFile := "apple2/romdumps/Apple2e.rom"

	log := false
	a := apple2.NewApple2(romFile)
	a.Run(log)
}
