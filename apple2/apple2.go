package apple2

import (
	"bufio"
	"fmt"
	"go6502/core6502"
	"os"
	"time"
)

// Apple2 represents all the components and state of the emulated machine
type Apple2 struct {
	cpu             *core6502.State
	mmu             *memoryManager
	io              *ioC0Page
	cg              *CharacterGenerator
	cards           []cardBase
	isApple2e       bool
	panicSS         bool
	activeSlot      int // Slot that has the addressing 0xc800 to 0ccfff
	commandChannel  chan int
	cycleDurationNs float64 // Inverse of the cpu clock in Ghz
	isColor         bool
}

const (
	// CpuClockMhz is the actual Apple II clock speed
	CpuClockMhz     = 14.318 / 14
	cpuClockEuroMhz = 14.238 / 14
)

// NewApple2 instantiates an apple2
func NewApple2(romFile string, charRomFile string, clockMhz float64, isColor bool, panicSS bool) *Apple2 {
	var a Apple2
	a.mmu = newMemoryManager(&a)
	a.cpu = core6502.NewNMOS6502(a.mmu)
	a.loadRom(romFile)
	if charRomFile != "" {
		a.cg = NewCharacterGenerator(charRomFile)
	}
	a.mmu.resetRomPaging()
	a.commandChannel = make(chan int, 100)
	a.isColor = isColor
	a.panicSS = panicSS

	if clockMhz <= 0 {
		// Full speed
		a.cycleDurationNs = 0
	} else {
		a.cycleDurationNs = 1000.0 / clockMhz
	}

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

const (
	// CommandToggleSpeed toggles cpu speed between full speed and actual Apple II speed
	CommandToggleSpeed = iota + 1
	// CommandToggleColor toggles between NTSC color TV and Green phospor monitor
	CommandToggleColor
)

// SendCommand enqueues a command to the emulator thread
func (a *Apple2) SendCommand(command int) {
	a.commandChannel <- command
}

func (a *Apple2) executeCommand(command int) {
	switch command {
	case CommandToggleSpeed:
		if a.cycleDurationNs == 0 {
			fmt.Println("Slow")
			a.cycleDurationNs = 1000.0 / CpuClockMhz
		} else {
			fmt.Println("Fast")
			a.cycleDurationNs = 0
		}
	case CommandToggleColor:
		a.isColor = !a.isColor
	}
}

// Run starts the Apple2 emulation
func (a *Apple2) Run(log bool) {
	// Start the processor
	a.cpu.Reset()
	referenceTime := time.Now()
	for {
		// Run a 6502 step
		a.cpu.ExecuteInstruction(log)

		// Execute meta commands
		commandsPending := true
		for commandsPending {
			select {
			case command := <-a.commandChannel:
				a.executeCommand(command)
			default:
				commandsPending = false
			}
		}

		if a.cycleDurationNs != 0 {
			// Wait until next 6502 step has to run
			clockDuration := time.Since(referenceTime)
			simulatedDuration := time.Duration(float64(a.cpu.GetCycles()) * a.cycleDurationNs)
			waitDuration := simulatedDuration - clockDuration
			if waitDuration > 1*time.Second {
				// We have to wait too long. Let's fast forward
				referenceTime = referenceTime.Add(-waitDuration)
				waitDuration = 0
			}
			if waitDuration > 0 {
				time.Sleep(waitDuration)
			}
		}
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
