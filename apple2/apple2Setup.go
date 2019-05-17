package apple2

import (
	"go6502/core6502"
)

// NewApple2 instantiates an apple2
func NewApple2(romFile string, charRomFile string, clockMhz float64,
	isColor bool, fastMode bool, panicSS bool) *Apple2 {
	var a Apple2
	a.persistance = newPersistance(&a)
	a.mmu = newMemoryManager(&a, romFile)
	a.persistance.register(a.mmu)
	a.cpu = core6502.NewNMOS6502(a.mmu)
	if charRomFile != "" {
		a.cg = NewCharacterGenerator(charRomFile)
	}
	a.commandChannel = make(chan int, 100)
	a.isColor = isColor
	a.fastMode = fastMode
	a.panicSS = panicSS

	if clockMhz <= 0 {
		// Full speed
		a.cycleDurationNs = 0
	} else {
		a.cycleDurationNs = 1000.0 / clockMhz
	}

	// Set the io in 0xc000
	a.io = newIoC0Page(&a)
	a.persistance.register(a.io)
	a.mmu.setPages(0xc0, 0xc0, a.io)

	return &a
}

// AddDisk2 insterts a DiskII controller
func (a *Apple2) AddDisk2(slot int, diskRomFile string, diskImage string) {
	d := newCardDisk2(diskRomFile)
	d.cardBase.insert(a, slot)
	a.persistance.register(d)

	if diskImage != "" {
		diskette := loadDisquette(diskImage)
		//diskette.saveNib(diskImage + "bak")
		d.drive[0].insertDiskette(diskette)
	}
}

// AddLanguageCard inserts a 16Kb card
func (a *Apple2) AddLanguageCard(slot int) {
	d := newCardLanguage()
	d.cardBase.insert(a, slot)
	d.applyState()
	a.persistance.register(d)
}

// AddSaturnCard inserts a 128Kb card
func (a *Apple2) AddSaturnCard(slot int) {
	d := newCardSaturn()
	d.cardBase.insert(a, slot)
	d.applyState()
	a.persistance.register(d)
}

// ConfigureStdConsole uses stdin and stdout to interface with the Apple2
func (a *Apple2) ConfigureStdConsole(stdinKeyboard bool, stdoutScreen bool) {
	if !stdinKeyboard && !stdoutScreen {
		return
	}

	// Init frontend
	fe := newAnsiConsoleFrontend(a, stdinKeyboard)
	if stdinKeyboard {
		a.io.setKeyboardProvider(fe)
	}
	if stdoutScreen {
		go fe.textModeGoRoutine()
	}
}

// SetKeyboardProvider attaches an external keyboard provider
func (a *Apple2) SetKeyboardProvider(kb KeyboardProvider) {
	a.io.setKeyboardProvider(kb)
}

// SetSpeakerProvider attaches an external keyboard provider
func (a *Apple2) SetSpeakerProvider(s SpeakerProvider) {
	a.io.setSpeakerProvider(s)
}
