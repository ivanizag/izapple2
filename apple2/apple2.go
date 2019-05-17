package apple2

import (
	"fmt"
	"go6502/core6502"
	"time"
)

// Apple2 represents all the components and state of the emulated machine
type Apple2 struct {
	cpu                 *core6502.State
	mmu                 *memoryManager
	io                  *ioC0Page
	cg                  *CharacterGenerator
	isApple2e           bool
	panicSS             bool
	commandChannel      chan int
	cycleDurationNs     float64 // Current speed. Inverse of the cpu clock in Ghz
	isColor             bool
	fastMode            bool
	fastRequestsCounter int
	persistance         *persistance
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
		a.persistance.save("apple2.state")
	case CommandLoadState:
		fmt.Println("Loading state")
		a.persistance.load("apple2.state")
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
