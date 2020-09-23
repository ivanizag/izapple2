package apple2

import (
	"errors"

	"github.com/ivanizag/apple2/core6502"
)

func newApple2() *Apple2 {
	var a Apple2

	a.Name = "Pending"
	a.mmu = newMemoryManager(&a)
	a.io = newIoC0Page(&a)
	return &a
}

func (a *Apple2) setup(isColor bool, clockMhz float64, fastMode bool, traceMLI bool) {
	a.commandChannel = make(chan int, 100)
	a.isColor = isColor
	a.fastMode = fastMode
	if traceMLI {
		a.traceMLI = newTraceProDOS(a)
	}

	if clockMhz <= 0 {
		// Full speed
		a.cycleDurationNs = 0
	} else {
		a.cycleDurationNs = 1000.0 / clockMhz
	}
}

func setApple2plus(a *Apple2) {
	a.Name = "Apple ][+"
	a.cpu = core6502.NewNMOS6502(a.mmu)
	addApple2SoftSwitches(a.io)
}

func setApple2e(a *Apple2) {
	a.Name = "Apple IIe"
	a.isApple2e = true
	a.cpu = core6502.NewNMOS6502(a.mmu)
	a.mmu.initExtendedRAM(1)
	addApple2SoftSwitches(a.io)
	addApple2ESoftSwitches(a.io)
}

func setApple2eEnhanced(a *Apple2) {
	a.Name = "Apple //e"
	a.isApple2e = true
	a.cpu = core6502.NewCMOS65c02(a.mmu)
	a.mmu.initExtendedRAM(1)
	addApple2SoftSwitches(a.io)
	addApple2ESoftSwitches(a.io)
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

	romBase := 0x10000 - size
	a.mmu.physicalROM[0] = newMemoryRangeROM(uint16(romBase), data, "Main ROM")
	return nil
}

// AddDisk2 inserts a DiskII controller
func (a *Apple2) AddDisk2(slot int, diskRomFile string, diskImage, diskBImage string) error {
	var c cardDisk2
	data, err := loadResource(diskRomFile)
	if err != nil {
		return err
	}
	c.loadRom(data)
	a.insertCard(&c, slot)

	if diskImage != "" {
		diskette, err := loadDiskette(diskImage)
		if err != nil {
			return err
		}
		c.drive[0].insertDiskette(diskette)
	}

	if diskBImage != "" {
		diskette, err := loadDiskette(diskBImage)
		if err != nil {
			return err
		}
		c.drive[1].insertDiskette(diskette)
	}

	return nil
}

// AddSmartPortDisk adds a smart port card and image
func (a *Apple2) AddSmartPortDisk(slot int, hdImage string, trace bool) error {
	var c cardHardDisk
	c.setTrace(trace)
	c.loadRom(buildHardDiskRom(slot))
	a.insertCard(&c, slot)

	hd, err := openBlockDisk(hdImage)
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

// AddMemoryExpansionCard inserts an Apple II Memory Expansion card with 1GB
func (a *Apple2) AddMemoryExpansionCard(slot int, romFile string) error {
	var c cardMemoryExpansion
	data, err := loadResource(romFile)
	if err != nil {
		return err
	}
	c.loadRom(data)
	a.insertCard(&c, slot)
	return nil
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

// AddRGBCard inserts an RBG option to the Apple IIe 80 col 64KB card
func (a *Apple2) AddRGBCard() {
	setupRGBCard(a)
}

// AddRAMWorks inserts adds RAMWorks style RAM to the Apple IIe 80 col 64KB card
func (a *Apple2) AddRAMWorks(banks int) {
	setupRAMWorksCard(a, banks)
}

// AddNoSlotClock inserts a DS1215 no slot clock under the main ROM
func (a *Apple2) AddNoSlotClock() {
	nsc := newNoSlotClockDS1216(a, a.mmu.physicalROM[0])
	a.mmu.physicalROM[0] = nsc
}

// AddNoSlotClockInCard inserts a DS1215 no slot clock under a card ROM
func (a *Apple2) AddNoSlotClockInCard(slot int) error {
	cardRom := a.mmu.cardsROM[slot]
	if cardRom == nil {
		return errors.New("No ROM available on the slot to add a no slot clock")
	}
	nsc := newNoSlotClockDS1216(a, cardRom)
	a.mmu.cardsROM[slot] = nsc
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
