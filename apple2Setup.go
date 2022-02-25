package izapple2

import (
	"errors"

	"github.com/ivanizag/iz6502"
)

func newApple2() *Apple2 {
	var a Apple2

	a.Name = "Pending"
	a.mmu = newMemoryManager(&a)
	a.io = newIoC0Page(&a)
	return &a
}

func (a *Apple2) setup(clockMhz float64, fastMode bool) {
	a.commandChannel = make(chan int, 100)
	a.fastMode = fastMode

	if clockMhz <= 0 {
		// Full speed
		a.cycleDurationNs = 0
	} else {
		a.cycleDurationNs = 1000.0 / clockMhz
	}
}

func setApple2plus(a *Apple2) {
	a.Name = "Apple ][+"
	a.cpu = iz6502.NewNMOS6502(a.mmu)
	addApple2SoftSwitches(a.io)
}

func setApple2e(a *Apple2) {
	a.Name = "Apple IIe"
	a.isApple2e = true
	a.cpu = iz6502.NewNMOS6502(a.mmu)
	a.mmu.initExtendedRAM(1)
	addApple2SoftSwitches(a.io)
	addApple2ESoftSwitches(a.io)
}

func setApple2eEnhanced(a *Apple2) {
	a.Name = "Apple //e"
	a.isApple2e = true
	a.cpu = iz6502.NewCMOS65c02(a.mmu)
	a.mmu.initExtendedRAM(1)
	addApple2SoftSwitches(a.io)
	addApple2ESoftSwitches(a.io)
}

func (a *Apple2) addTracer(tracer executionTracer) {
	if a.tracers == nil {
		a.tracers = make([]executionTracer, 0)
	}

	a.tracers = append(a.tracers, tracer)
}

func (a *Apple2) insertCard(c Card, slot int) {
	c.assign(a, slot)
	a.cards[slot] = c
}

// GetCards returns the array of inserted cards
func (a *Apple2) GetCards() [8]Card {
	return a.cards
}

const (
	apple2RomSize  = 12 * 1024
	apple2eRomSize = 16 * 1024
)

// LoadRom loads a standard Apple2+ or 2e ROM
func (a *Apple2) LoadRom(filename string) error {
	data, _, err := LoadResource(filename)
	if err != nil {
		return err
	}

	size := len(data)
	if size != apple2RomSize && size != apple2eRomSize {
		return errors.New("rom size not supported")
	}

	romBase := 0x10000 - size
	a.mmu.physicalROM[0] = newMemoryRangeROM(uint16(romBase), data, "Main ROM")
	return nil
}

// AddDisk2 inserts a DiskII controller
func (a *Apple2) AddDisk2(slot int, diskImage, diskBImage string, trackTracer trackTracer) error {
	c := NewCardDisk2(trackTracer)
	a.insertCard(c, slot)

	if diskImage != "" {
		diskette, err := LoadDiskette(diskImage)
		if err != nil {
			return err
		}
		c.drive[0].insertDiskette(diskImage, diskette)
	}

	if diskBImage != "" {
		diskette, err := LoadDiskette(diskBImage)
		if err != nil {
			return err
		}
		c.drive[1].insertDiskette(diskImage, diskette)
	}

	return nil
}

// AddDisk2 inserts a DiskII controller
func (a *Apple2) AddDisk2Sequencer(slot int, diskImage, diskBImage string, trackTracer trackTracer) error {
	c := NewCardDisk2Sequencer(trackTracer)
	a.insertCard(c, slot)

	if diskImage != "" {
		err := c.drive[0].insertDiskette(diskImage)
		if err != nil {
			return err
		}
	}

	if diskBImage != "" {
		err := c.drive[0].insertDiskette(diskBImage)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddSmartPortDisk adds a smart port card and image
func (a *Apple2) AddSmartPortDisk(slot int, hdImage string, trace bool) error {
	c := NewCardHardDisk()
	c.trace = trace
	err := c.LoadImage(hdImage)
	if err != nil {
		return err
	}
	a.insertCard(c, slot)
	return nil
}

// AddVidHD adds a card with the signature of VidHD
func (a *Apple2) AddVidHD(slot int) {
	a.insertCard(NewCardVidHD(), slot)
}

// AddFastChip adds a card with the signature of VidHD
func (a *Apple2) AddFastChip(slot int) {
	a.insertCard(NewCardFastChip(), slot)
}

// AddLanguageCard inserts a 16Kb card
func (a *Apple2) AddLanguageCard(slot int) {
	a.insertCard(NewCardLanguage(), slot)
}

// AddSaturnCard inserts a 128Kb card
func (a *Apple2) AddSaturnCard(slot int) {
	a.insertCard(NewCardSaturn(), slot)
}

// AddMemoryExpansionCard inserts an Apple II Memory Expansion card with 1GB
func (a *Apple2) AddMemoryExpansionCard(slot int) {
	a.insertCard(NewCardMemoryExpansion(), slot)
}

// AddThunderClockPlusCard inserts a ThunderClock Plus clock card
func (a *Apple2) AddThunderClockPlusCard(slot int) {
	a.insertCard(NewCardThunderClockPlus(), slot)
}

// AddMouseCard inserts a Mouse card
func (a *Apple2) AddMouseCard(slot int) {
	a.insertCard(NewCardMouse(), slot)
}

// AddVidexCard inserts a Videx card
func (a *Apple2) AddVidexCard(slot int) {
	c := NewCardVidex()
	a.insertCard(c, slot)
	a.softVideoSwitch = NewSoftVideoSwitch(c)
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

// AddRomX inserts a RomX. It intercepts all memory accesses
func (a *Apple2) AddRomX() error {
	rx, err := newRomX(a, a.mmu)
	if err != nil {
		return err
	}
	a.cpu.SetMemory(rx)
	return nil
}

// AddNoSlotClockInCard inserts a DS1215 no slot clock under a card ROM
func (a *Apple2) AddNoSlotClockInCard(slot int) error {
	cardRom := a.mmu.cardsROM[slot]
	if cardRom == nil {
		return errors.New("no ROM available on the slot to add a no slot clock")
	}
	nsc := newNoSlotClockDS1216(a, cardRom)
	a.mmu.cardsROM[slot] = nsc
	return nil
}

// AddCardLogger inserts a fake card that logs accesses
func (a *Apple2) AddCardLogger(slot int) {
	c := NewCardLogger()
	a.insertCard(c, slot)
}

// AddCardInOut inserts a fake card that interfaces with the emulator host
func (a *Apple2) AddCardInOut(slot int) {
	c := NewCardInOut()
	a.insertCard(c, slot)
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

// SetMouseProvider attaches an external joysticks provider
func (a *Apple2) SetMouseProvider(m MouseProvider) {
	a.io.setMouseProvider(m)
}
