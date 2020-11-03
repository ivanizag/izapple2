package izapple2

import (
	"flag"
	"os"

	"github.com/ivanizag/izapple2/storage"
)

const defaultInternal = "<default>"

// MainApple is a device independent main. Video, keyboard and speaker won't be defined
func MainApple() *Apple2 {
	romFile := flag.String(
		"rom",
		defaultInternal,
		"main rom file")
	disk2Slot := flag.Int(
		"disk2Slot",
		6,
		"slot for the disk driver. -1 for none.")
	diskImage := flag.String(
		"disk",
		"<internal>/dos33.dsk",
		"file to load on the first disk drive")
	diskBImage := flag.String(
		"diskb",
		"",
		"file to load on the second disk drive")
	hardDiskImage := flag.String(
		"hd",
		"",
		"file to load on the boot hard disk (slot 7)")
	hardDiskSlot := flag.Int(
		"hdSlot",
		-1,
		"slot for the hard drive if present. -1 for none.")
	smartPortImage := flag.String(
		"disk35",
		"",
		"file to load on the SmartPort disk (slot 5)")
	cpuClock := flag.Float64(
		"mhz",
		CPUClockMhz,
		"cpu speed in Mhz, use 0 for full speed. Use F5 to toggle.")
	charRomFile := flag.String(
		"charRom",
		defaultInternal,
		"rom file for the character generator")
	languageCardSlot := flag.Int(
		"languageCardSlot",
		0,
		"slot for the 16kb language card. -1 for none")
	saturnCardSlot := flag.Int(
		"saturnCardSlot",
		-1,
		"slot for the 256kb Saturn card. -1 for none")
	vidHDCardSlot := flag.Int(
		"vidHDSlot",
		2,
		"slot for the VidHD card, only for //e models. -1 for none")
	fastChipCardSlot := flag.Int(
		"fastChipSlot",
		3,
		"slot for the FASTChip accelerator card, -1 for none")
	memoryExpansionCardSlot := flag.Int(
		"memoryExpSlot",
		1,
		"slot for the Memory Expansion card with 1GB. -1 for none")
	ramWorksKb := flag.Int(
		"ramworks",
		8192,
		"memory to use with RAMWorks card, 0 for no card, max is 16384")
	thunderClockCardSlot := flag.Int(
		"thunderClockCardSlot",
		4,
		"slot for the ThunderClock Plus card. -1 for none")
	nsc := flag.Int(
		"nsc",
		-1,
		"add a DS1216 No-Slot-Clock on the main ROM (use 0) or a slot ROM. -1 for none")
	rgbCard := flag.Bool(
		"rgb",
		true,
		"emulate the RGB modes of the 80col RGB card for DHGR")
	fastDisk := flag.Bool(
		"fastDisk",
		true,
		"set fast mode when the disks are spinning")
	panicSS := flag.Bool(
		"panicSS",
		false,
		"panic if a not implemented softswitch is used")
	traceCPU := flag.Bool(
		"traceCpu",
		false,
		"dump to the console the CPU execution operations")
	traceSS := flag.Bool(
		"traceSS",
		false,
		"dump to the console the sofswitches calls")
	traceSSReg := flag.Bool(
		"traceSSReg",
		false,
		"dump to the console the sofswitch registrations")
	traceHD := flag.Bool(
		"traceHD",
		false,
		"dump to the console the hd/smarport commands")
	dumpChars := flag.Bool(
		"dumpChars",
		false,
		"shows the character map")
	model := flag.String(
		"model",
		"2enh",
		"set base model. Models available 2plus, 2e, 2enh, base64a")
	profile := flag.Bool(
		"profile",
		false,
		"generate profile trace to analyse with pprof")
	traceMLI := flag.Bool(
		"traceMLI",
		false,
		"dump to the console the calls to ProDOS machine language interface calls to $BF00")
	forceCaps := flag.Bool(
		"forceCaps",
		false,
		"force all letters to be uppercased (no need for caps lock!)")

	flag.Parse()

	// Process a filename with autodetection
	filename := flag.Arg(0)
	diskImageFinal := *diskImage
	hardDiskImageFinal := *hardDiskImage
	if filename != "" {
		if storage.IsDiskette(filename) {
			diskImageFinal = filename
		} else {
			hardDiskImageFinal = filename
		}
	}

	a := newApple2()
	a.setup(*cpuClock, *fastDisk, *traceMLI)
	a.io.setTrace(*traceSS)
	a.io.setTraceRegistrations(*traceSSReg)
	a.io.setPanicNotImplemented(*panicSS)
	a.setProfiling(*profile)
	a.SetForceCaps(*forceCaps)

	var charGenMap charColumnMap
	initialCharGenPage := 0
	switch *model {
	case "2plus":
		setApple2plus(a)
		if *romFile == defaultInternal {
			*romFile = "<internal>/Apple2_Plus.rom"
		}
		if *charRomFile == defaultInternal {
			*charRomFile = "<internal>/Apple2rev7CharGen.rom"
		}
		charGenMap = charGenColumnsMap2Plus
		*vidHDCardSlot = -1

	case "2e":
		setApple2e(a)
		if *romFile == defaultInternal {
			*romFile = "<internal>/Apple2e.rom"
		}
		if *charRomFile == defaultInternal {
			*charRomFile = "<internal>/Apple IIe Video Unenhanced - 342-0133-A - 2732.bin"
		}
		a.isApple2e = true
		charGenMap = charGenColumnsMap2e

	case "2enh":
		setApple2eEnhanced(a)
		if *romFile == defaultInternal {
			*romFile = "<internal>/Apple2e_Enhanced.rom"
		}
		if *charRomFile == defaultInternal {
			*charRomFile = "<internal>/Apple IIe Video Enhanced - 342-0265-A - 2732.bin"
		}
		a.isApple2e = true
		charGenMap = charGenColumnsMap2e

	case "base64a":
		setBase64a(a)
		if *romFile == defaultInternal {
			err := loadBase64aRom(a)
			if err != nil {
				panic(err)
			}
			*romFile = ""
		}
		if *charRomFile == defaultInternal {
			*charRomFile = "<internal>/BASE64A_ROM7_CharGen.BIN"
			initialCharGenPage = 1
		}
		charGenMap = charGenColumnsMapBase64a
		*vidHDCardSlot = -1

	default:
		panic("Model not supported")
	}

	a.cpu.SetTrace(*traceCPU)

	// Load ROM if not loaded already
	if *romFile != "" {
		err := a.LoadRom(*romFile)
		if err != nil {
			panic(err)
		}
	}

	// Load character generator if it loaded already
	cg, err := newCharacterGenerator(*charRomFile, charGenMap)
	if err != nil {
		panic(err)
	}
	cg.setPage(initialCharGenPage)
	a.cg = cg

	// Externsion cards
	if *languageCardSlot >= 0 {
		a.AddLanguageCard(*languageCardSlot)
	}
	if *saturnCardSlot >= 0 {
		a.AddSaturnCard(*saturnCardSlot)
	}
	if *memoryExpansionCardSlot >= 0 {
		a.AddMemoryExpansionCard(*memoryExpansionCardSlot)
	}
	if *thunderClockCardSlot > 0 {
		err := a.AddThunderClockPlusCard(*thunderClockCardSlot,
			"<internal>/ThunderclockPlusROM.bin")
		if err != nil {
			panic(err)
		}
	}
	if *vidHDCardSlot >= 0 {
		a.AddVidHD(*vidHDCardSlot)
	}

	if *smartPortImage != "" {
		err := a.AddSmartPortDisk(5, *smartPortImage, *traceHD)
		if err != nil {
			panic(err)
		}
	}

	if *fastChipCardSlot >= 0 {
		a.AddFastChip(*fastChipCardSlot)
	}
	if *disk2Slot > 0 {
		err := a.AddDisk2(*disk2Slot, diskImageFinal, *diskBImage)
		if err != nil {
			panic(err)
		}
	}
	if hardDiskImageFinal != "" {
		if *hardDiskSlot <= 0 {
			// If there is a hard disk image, but no slot assigned, use slot 7.
			*hardDiskSlot = 7
		}
		err := a.AddSmartPortDisk(*hardDiskSlot, hardDiskImageFinal, *traceHD)
		if err != nil {
			panic(err)
		}
	}

	if *ramWorksKb != 0 {
		if *ramWorksKb%64 != 0 {
			panic("Ramworks size must be a multiple of 64")
		}
		a.AddRAMWorks(*ramWorksKb / 64)
	}

	if *rgbCard {
		a.AddRGBCard()
	}

	if *nsc == 0 {
		a.AddNoSlotClock()
	} else if *nsc > 0 {
		err := a.AddNoSlotClockInCard(*nsc)
		if err != nil {
			panic(err)
		}

	}

	// a.AddRomX()
	// a.AddCardInOut(2)
	// a.AddCardLogger(4)

	if *dumpChars {
		a.cg.Dump()
		os.Exit(0)
		return nil
	}

	return a
}
