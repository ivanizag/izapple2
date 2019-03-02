package apple2

import (
	"bufio"
	"go6502/core6502"
	"os"
)

type Apple2 struct {
	cpu        *core6502.State
	mmu        *memoryManager
	isApple2e  bool
	ioPage     *ioC0Page // 0xc000 to 0xc080
	activeSlot int       // Slot that has the addressing 0xc800 to 0ccfff
}

// NewApple2 instantiates an apple2
func NewApple2(romFile string) *Apple2 {
	var a Apple2
	a.mmu = newMemoryManager(&a)
	a.cpu = core6502.NewNMOS6502(a.mmu)
	a.loadRom(romFile)
	a.mmu.resetPaging()

	// Set the io in 0xc000
	a.ioPage = newIoC0Page(&a)
	a.mmu.setPage(0xc0, a.ioPage)

	return &a
}

func (a *Apple2) Run(log bool) {
	// Init frontend
	fe := newAnsiConsoleFrontend(a)
	a.ioPage.setKeyboardProvider(fe)
	go fe.textModeGoRoutine()

	// Start the processor
	a.cpu.Reset()
	for {
		a.cpu.ExecuteInstruction(log)
	}
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
