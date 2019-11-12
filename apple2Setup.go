package apple2

import (
	"errors"

	"github.com/ivanizag/apple2/core6502"
)

// newApple2 instantiates an apple2
func newApple2plus() *Apple2 {
	var a Apple2
	a.Name = "Apple ][+"
	a.mmu = newMemoryManager(&a)
	a.cpu = core6502.NewNMOS6502(a.mmu)
	a.io = newIoC0Page(&a)
	addApple2SoftSwitches(a.io)

	return &a
}

func newApple2e() *Apple2 {
	var a Apple2
	a.Name = "Apple IIe"
	a.mmu = newMemoryManager(&a)
	a.cpu = core6502.NewNMOS6502(a.mmu)
	a.io = newIoC0Page(&a)
	a.mmu.InitRAMalt()
	addApple2SoftSwitches(a.io)
	addApple2ESoftSwitches(a.io)

	return &a
}

func newApple2eEnhanced() *Apple2 {
	var a Apple2
	a.Name = "Apple //e"
	a.mmu = newMemoryManager(&a)
	a.cpu = core6502.NewCMOS65c02(a.mmu)
	a.io = newIoC0Page(&a)
	a.mmu.InitRAMalt()
	addApple2SoftSwitches(a.io)
	addApple2ESoftSwitches(a.io)

	return &a
}

func (a *Apple2) setup(isColor bool, clockMhz float64, fastMode bool) {
	a.commandChannel = make(chan int, 100)
	a.isColor = isColor
	a.fastMode = fastMode

	if clockMhz <= 0 {
		// Full speed
		a.cycleDurationNs = 0
	} else {
		a.cycleDurationNs = 1000.0 / clockMhz
	}
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
func (a *Apple2) LoadRom(filename string) error {
	data, err := loadResource(filename)
	if err != nil {
		return err
	}

	size := len(data)
	if size != apple2RomSize && size != apple2eRomSize {
		return errors.New("Rom size not supported")
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

	mmu.physicalROM[0] = newMemoryRange(0xd000, data[romStart:])
	return nil
}

// AddDisk2 inserts a DiskII controller
func (a *Apple2) AddDisk2(slot int, diskRomFile string, diskImage string) error {
	var c cardDisk2
	data, err := loadResource(diskRomFile)
	if err != nil {
		return err
	}
	c.loadRom(data)
	a.insertCard(&c, slot)

	if diskImage != "" {
		diskette, err := loadDisquette(diskImage)
		if err != nil {
			return err
		}
		c.drive[0].insertDiskette(diskette)
	}

	return nil
}

// AddHardDisk adds a ProDos hard dirve with a 2MG image
func (a *Apple2) AddHardDisk(slot int, hdImage string, trace bool) error {
	var c cardHardDisk
	c.setTrace(trace)
	c.loadRom(buildHardDiskRom(slot))
	a.insertCard(&c, slot)

	hd, err := openHardDisk2mg(hdImage)
	if err != nil {
		return err
	}
	c.addDisk(hd)
	return nil
}

// AddVidHD adds a card with the signature of VidHD
func (a *Apple2) AddVidHD(slot int) {
	var c cardVidHD
	c.loadRom(buildVidHDRom())
	a.insertCard(&c, slot)
}

// AddFastChip adds a card with the signature of VidHD
func (a *Apple2) AddFastChip(slot int) {
	var c cardFastChip
	c.loadRom(buildFastChipRom())
	a.insertCard(&c, slot)
}

// AddLanguageCard inserts a 16Kb card
func (a *Apple2) AddLanguageCard(slot int) {
	a.insertCard(&cardLanguage{}, slot)
}

// AddSaturnCard inserts a 128Kb card
func (a *Apple2) AddSaturnCard(slot int) {
	a.insertCard(&cardSaturn{}, slot)
}

// AddThunderClockPlusCard inserts a ThunderClock Plus clock card
func (a *Apple2) AddThunderClockPlusCard(slot int, romFile string) error {
	var c cardThunderClockPlus
	data, err := loadResource(romFile)
	if err != nil {
		return err
	}
	c.loadRom(data)
	a.insertCard(&c, slot)
	return nil
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
