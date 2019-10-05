package apple2

import (
	"flag"
	"os"
)

// MainApple is a device independant main. Video, keyboard and speaker won't be defined
func MainApple() *Apple2 {
	romFile := flag.String(
		"rom",
		"<internal>/Apple2_Plus.rom",
		"main rom file")
	disk2RomFile := flag.String(
		"diskRom",
		"<internal>/DISK2.rom",
		"rom file for the disk drive controller")
	disk2Slot := flag.Int(
		"disk2Slot",
		6,
		"slot for the disk driver. -1 for none.")
	diskImage := flag.String(
		"disk",
		"<internal>/dos33.dsk",
		"file to load on the first disk drive")
	hardDiskImage := flag.String(
		"hd",
		"",
		"file to load on the hard disk")
	hardDiskSlot := flag.Int(
		"hdSlot",
		-1,
		"slot for the hard drive if present. -1 for none.")
	cpuClock := flag.Float64(
		"mhz",
		CpuClockMhz,
		"cpu speed in Mhz, use 0 for full speed. Use F5 to toggle.")
	charRomFile := flag.String(
		"charRom",
		"<internal>/Apple2rev7CharGen.rom",
		"rom file for the character generator")
	languageCardSlot := flag.Int(
		"languageCardSlot",
		0,
		"slot for the 16kb language card. -1 for none")
	saturnCardSlot := flag.Int(
		"saturnCardSlot",
		-1,
		"slot for the 256kb Saturn card. -1 for none")
	thunderClockCardSlot := flag.Int(
		"thunderClockCardSlot",
		4,
		"slot for the ThunderClock Plus card. -1 for none")
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
		"panic if a not implemented softswitch is used",
	)
	traceCPU := flag.Bool(
		"traceCpu",
		false,
		"dump to the console the CPU execution operations",
	)
	traceSS := flag.Bool(
		"traceSS",
		false,
		"dump to the console the sofswitches calls",
	)
	dumpChars := flag.Bool(
		"dumpChars",
		false,
		"shows the character map",
	)
	base64a := flag.Bool(
		"base64a",
		false,
		"setup a Base64A clone",
	)
	flag.Parse()

	a := NewApple2(*cpuClock, !*mono, *fastDisk)

	a.cpu.SetTrace(*traceCPU)
	a.io.setTrace(*traceSS)
	a.io.setPanicNotImplemented(*panicSS)

	if *charRomFile != "" {
		cg, err := NewCharacterGenerator(*charRomFile)
		if err != nil {
			panic(err)
		}
		a.cg = cg
	}

	if *base64a {
		NewBase64a(a)
	} else {
		err := a.LoadRom(*romFile)
		if err != nil {
			panic(err)
		}
	}
	if *languageCardSlot >= 0 {
		a.AddLanguageCard(*languageCardSlot)
	}
	if *saturnCardSlot >= 0 {
		a.AddSaturnCard(*saturnCardSlot)
	}
	if *thunderClockCardSlot > 0 {
		err := a.AddThunderClockPlusCard(*thunderClockCardSlot, "<internal>/ThunderclockPlusROM.bin")
		if err != nil {
			panic(err)
		}
	}
	if *disk2Slot > 0 {
		err := a.AddDisk2(*disk2Slot, *disk2RomFile, *diskImage)
		if err != nil {
			panic(err)
		}
	}
	if *hardDiskImage != "" {
		if *hardDiskSlot <= 0 {
			// If there is a hard disk image, but no slot assigned, use slot 7.
			*hardDiskSlot = 7
		}
		err := a.AddHardDisk(*hardDiskSlot, *hardDiskImage)
		if err != nil {
			panic(err)
		}
	}

	//a.AddCardInOut(2)
	//a.AddCardLogger(4)

	if *dumpChars {
		a.cg.Dump()
		os.Exit(0)
		return nil
	}

	return a
}
