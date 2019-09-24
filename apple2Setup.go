package apple2

import "github.com/ivanizag/apple2/core6502"

// NewApple2 instantiates an apple2
func NewApple2(charRomFile string, clockMhz float64,
	isColor bool, fastMode bool) *Apple2 {

	var a Apple2
	a.Name = "Apple ][+"
	a.mmu = newMemoryManager(&a)
	a.cpu = core6502.NewNMOS6502(a.mmu)
	if charRomFile != "" {
		a.cg = NewCharacterGenerator(charRomFile)
	}
	a.commandChannel = make(chan int, 100)
	a.isColor = isColor
	a.fastMode = fastMode

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

const (
	apple2RomSize  = 12 * 1024
	apple2eRomSize = 16 * 1024
)

// LoadRom loads a standard Apple2+ or 2e ROM
func (a *Apple2) LoadRom(filename string) {
	data := loadResource(filename)
	size := len(data)
	if size != apple2RomSize && size != apple2eRomSize {
		panic("Rom size not supported")
	}

	romStart := 0
	mmu := a.mmu
	if size == apple2eRomSize {
		// The extra 4kb ROM is first in the rom file.
		// It starts with 256 unused bytes not mapped to 0xc000.
		a.isApple2e = true
		extraRomSize := apple2eRomSize - apple2RomSize
		mmu.physicalROMe = newMemoryRange(0xc000, data[0:extraRomSize])
		romStart = extraRomSize
	}

	mmu.physicalROM = newMemoryRange(0xd000, data[romStart:])
	mmu.resetRomPaging()
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

// SetJoysticksProvider attaches an external joysticks provider
func (a *Apple2) SetJoysticksProvider(j JoysticksProvider) {
	a.io.setJoysticksProvider(j)
}
