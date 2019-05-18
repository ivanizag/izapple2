package apple2

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"go6502/core6502"
	"io"
	"os"
	"time"
)

// Apple2 represents all the components and state of the emulated machine
type Apple2 struct {
	cpu                 *core6502.State
	mmu                 *memoryManager
	io                  *ioC0Page
	cg                  *CharacterGenerator
	cards               [8]card
	isApple2e           bool
	panicSS             bool
	commandChannel      chan int
	cycleDurationNs     float64 // Current speed. Inverse of the cpu clock in Ghz
	isColor             bool
	fastMode            bool
	fastRequestsCounter int
}

const (
	// CpuClockMhz is the actual Apple II clock speed
	CpuClockMhz     = 14.318 / 14
	cpuClockEuroMhz = 14.238 / 14
)

const maxWaitDuration = 100 * time.Millisecond

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

		if a.cycleDurationNs != 0 && a.fastRequestsCounter <= 0 {
			// Wait until next 6502 step has to run
			clockDuration := time.Since(referenceTime)
			simulatedDuration := time.Duration(float64(a.cpu.GetCycles()) * a.cycleDurationNs)
			waitDuration := simulatedDuration - clockDuration
			if waitDuration > maxWaitDuration || -waitDuration > maxWaitDuration {
				// We have to wait too long or are too much behind. Let's fast forward
				referenceTime = referenceTime.Add(-waitDuration)
				waitDuration = 0
			}
			if waitDuration > 0 {
				time.Sleep(waitDuration)
			}
		}
	}
}

const (
	// CommandToggleSpeed toggles cpu speed between full speed and actual Apple II speed
	CommandToggleSpeed = iota + 1
	// CommandToggleColor toggles between NTSC color TV and Green phospor monitor
	CommandToggleColor
	// CommandSaveState stores the state to file
	CommandSaveState
	// CommandLoadState reload the last state
	CommandLoadState
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
	case CommandSaveState:
		fmt.Println("Saving state")
		a.save("apple2.state")
	case CommandLoadState:
		fmt.Println("Loading state")
		a.load("apple2.state")
	}
}

func (a *Apple2) requestFastMode() {
	// Note: if the fastMode is shorter than maxWaitDuration, there won't be any gain.
	if a.fastMode {
		a.fastRequestsCounter++
	}
}

func (a *Apple2) releaseFastMode() {
	if a.fastMode {
		a.fastRequestsCounter--
	}
}

type persistent interface {
	save(io.Writer)
	load(io.Reader)
}

func (a *Apple2) save(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	a.cpu.Save(w)
	a.mmu.save(w)
	a.io.save(w)
	binary.Write(w, binary.BigEndian, a.isColor)
	binary.Write(w, binary.BigEndian, a.fastMode)
	binary.Write(w, binary.BigEndian, a.fastRequestsCounter)

	for _, c := range a.cards {
		if c != nil {
			c.save(w)
		}
	}
}

func (a *Apple2) load(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		// Ignore error if can't load the file
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)

	a.cpu.Load(r)
	a.mmu.load(r)
	a.io.load(r)
	binary.Read(r, binary.BigEndian, &a.isColor)
	binary.Read(r, binary.BigEndian, &a.fastMode)
	binary.Read(r, binary.BigEndian, &a.fastRequestsCounter)

	for _, c := range a.cards {
		if c != nil {
			c.load(r)

		}
	}
}
