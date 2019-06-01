package apple2

import "github.com/ivanizag/apple2/core6502"

// NewApple2 instantiates an apple2
func NewApple2(romFile string, charRomFile string, clockMhz float64,
	isColor bool, fastMode bool, panicSS bool) *Apple2 {
	var a Apple2
	a.mmu = newMemoryManager(&a, romFile)
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
	a.mmu.setPages(0xc0, 0xc0, a.io)

	return &a
}

func (a *Apple2) insertCard(c card, slot int) {
	c.assign(a, slot)
	a.cards[slot] = c
}

// AddDisk2 insterts a DiskII controller
func (a *Apple2) AddDisk2(slot int, diskRomFile string, diskImage string) {
	var c cardDisk2
	c.loadRom(diskRomFile)
	a.insertCard(&c, slot)

	if diskImage != "" {
		diskette := loadDisquette(diskImage)
		//diskette.saveNib(diskImage + "bak")
		c.drive[0].insertDiskette(diskette)
	}
}

// AddLanguageCard inserts a 16Kb card
func (a *Apple2) AddLanguageCard(slot int) {
	a.insertCard(&cardLanguage{}, slot)
}

// AddSaturnCard inserts a 128Kb card
func (a *Apple2) AddSaturnCard(slot int) {
	a.insertCard(&cardSaturn{}, slot)
}

// AddCardLogger inserts a fake card that logs accesses
func (a *Apple2) AddCardLogger(slot int) {
	a.insertCard(&cardLogger{}, slot)
}

// AddCardInOut inserts a fake card that interfaces with the emulator host
func (a *Apple2) AddCardInOut(slot int) {
	a.insertCard(&cardInOut{}, slot)
}

// SetKeyboardProvider attaches an external keyboard provider
func (a *Apple2) SetKeyboardProvider(kb KeyboardProvider) {
	a.io.setKeyboardProvider(kb)
}

// SetSpeakerProvider attaches an external keyboard provider
func (a *Apple2) SetSpeakerProvider(s SpeakerProvider) {
	a.io.setSpeakerProvider(s)
}
