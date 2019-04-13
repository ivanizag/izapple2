package apple2

import (
	"bufio"
	"go6502/core6502"
	"os"
)

// Apple2 represents all the components and state of the emulated machine
type Apple2 struct {
	cpu        *core6502.State
	mmu        *memoryManager
	io         *ioC0Page
	cards      []cardBase
	isApple2e  bool
	activeSlot int // Slot that has the addressing 0xc800 to 0ccfff
}

// NewApple2 instantiates an apple2
func NewApple2(romFile string) *Apple2 {
	var a Apple2
	a.mmu = newMemoryManager(&a)
	a.cpu = core6502.NewNMOS6502(a.mmu)
	a.loadRom(romFile)
	a.mmu.resetPaging()

	// Set the io in 0xc000
	a.io = newIoC0Page(&a)
	a.mmu.setPage(0xc0, a.io)

	return &a
}

// AddDisk2 insterts a DiskII controller on slot 6
func (a *Apple2) AddDisk2(diskRomFile string, diskImage string) {
	d := newCardDisk2(diskRomFile)
	d.cardBase.insert(a, 6)

	if diskImage != "" {
		diskette := loadDisquette(diskImage)
		//diskette.saveNib(diskImage + "bak")
		d.drive[0].insertDiskette(diskette)
	}
}

// Run starts the Apple2 emulation
func (a *Apple2) Run(log bool, consoleKeyboard bool) {
	// Init frontend
	fe := newAnsiConsoleFrontend(a)
	if consoleKeyboard {
		a.io.setKeyboardProvider(fe)
	}
	if !log {
		go fe.textModeGoRoutine()
	}

	// Start the processor
	a.cpu.Reset()
	for {
		a.cpu.ExecuteInstruction(log)
	}
}

// SetKeyboardProvider attaches an external keyboard provider
func (a *Apple2) SetKeyboardProvider(kb KeyboardProvider) {
	a.io.setKeyboardProvider(kb)
}

// LoadRom loads a binary file to the top of the memory.
const (
	apple2RomSize  = 12 * 1024
	apple2eRomSize = 16 * 1024
)

func (a *Apple2) loadRom(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	stats, statsErr := f.Stat()
	if statsErr != nil {
		panic(err)
	}

	size := stats.Size()
	if size != apple2RomSize && size != apple2eRomSize {
		panic("Rom size not supported")
	}
	bytes := make([]byte, size)
	buf := bufio.NewReader(f)
	buf.Read(bytes)

	romStart := 0
	if size == apple2eRomSize {
		// The extra 4kb ROM is first in the rom file.
		// It starts with 256 unused bytes not mapped to 0xc000.
		a.isApple2e = true
		extraRomSize := apple2eRomSize - apple2RomSize
		a.mmu.physicalROMe = make([]romPage, extraRomSize>>8)
		for i := 0; i < extraRomSize; i++ {
			a.mmu.physicalROMe[i>>8].burn(uint8(i), bytes[i])
		}
		romStart = extraRomSize
	}

	a.mmu.physicalROM = make([]romPage, apple2RomSize>>8)
	for i := 0; i < apple2RomSize; i++ {
		a.mmu.physicalROM[i>>8].burn(uint8(i), bytes[i+romStart])
	}
}
