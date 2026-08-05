package izapple2

import (
	"fmt"
	"time"
)

const (
	// CPUClockMhz is the actual Apple II clock speed
	CPUClockMhz     = 14.318 / 14
	cpuClockEuroMhz = 14.238 / 14
)

const (
	maxWaitDuration = 100 * time.Millisecond
	cpuSpinLoops    = 100
)

// commandOutcome tells the run loop what a drained command requires from it
type commandOutcome int

const (
	// commandOutcomeNone when the run loop can continue as usual
	commandOutcomeNone commandOutcome = iota
	// commandOutcomeKill when the cpu execution loop must stop
	commandOutcomeKill
	// commandOutcomeResync when the emulation clock reference must be reset
	commandOutcomeResync
)

// Run starts the Apple2 emulation
func (a *Apple2) Run() {
	a.Start(false)
}

// Init resets the processor and prepares the machine to be run step by
// step with RunCycles. It is not needed with Run or Start, that do it
// themselves.
func (a *Apple2) Init() {
	a.cpu.Reset()
	a.cycles = a.cpu.GetCycles()
}

// RunCycles advances the emulation by at least the requested number of CPU
// cycles and returns the cycles actually executed, that can be a few more as
// the last instruction is not interrupted. It runs at full speed without the
// wall clock throttle of Start: the caller sets the pace. This is what the
// frontends that own the frame timing, like the libretro core, use instead of
// Start. Init must be called once before the first call, and all the calls
// must be made from the same goroutine.
func (a *Apple2) RunCycles(cycles uint64) uint64 {
	a.drainCommands()

	start := a.cycles
	for a.cycles-start < cycles {
		a.stepInstruction()
	}
	return a.cycles - start
}

// Start the Apple2 emulation, can start paused
func (a *Apple2) Start(paused bool) {
	// Start the processor
	a.Init()

	referenceTime := time.Now()
	speedReferenceTime := referenceTime
	speedReferenceCycles := uint64(0)

	a.paused.Store(paused)

	for {
		// Run cpu steps
		if !a.paused.Load() {
			for i := 0; i < cpuSpinLoops; i++ {
				a.stepInstruction()
			}

			if bp := a.cycleBreakpoint.Load(); bp != 0 && a.cycles >= bp {
				a.breakPoint.Store(true)
				a.cycleBreakpoint.Store(0)
				a.paused.Store(true)
			}
		} else {
			time.Sleep(200 * time.Millisecond)
		}

		// Execute meta commands
		switch a.drainCommands() {
		case commandOutcomeKill:
			return
		case commandOutcomeResync:
			referenceTime = time.Now()
			speedReferenceTime = referenceTime
		}

		if a.cycleDurationNs != 0 && a.fastRequestsCounter <= 0 {
			// Wait until next 6502 step has to run
			clockDuration := time.Since(referenceTime)
			simulatedDuration := time.Duration(float64(a.cycles) * a.cycleDurationNs)
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

		if a.cycles-speedReferenceCycles > 1000000 {
			// Calculate speed in MHz every million cycles
			newTime := time.Now()
			elapsedCycles := float64(a.cycles - speedReferenceCycles)
			a.currentFreqMHz = 1000.0 * elapsedCycles / float64(newTime.Sub(speedReferenceTime).Nanoseconds())
			speedReferenceTime = newTime
			speedReferenceCycles = a.cycles
		}
	}
}

// stepInstruction executes the next 6502 instruction, or a single DMA cycle
// when a card, like the Z80 Softcard, has taken over the bus
func (a *Apple2) stepInstruction() {
	if a.dmaActive {
		a.cards[a.dmaSlot].runDMACycle()
		a.cycles++
	} else {
		// Conditional tracing
		// pc, _ := a.cpu.GetPCAndSP()
		// a.cpu.SetTrace(pc >= 0xc700 && pc < 0xc800)

		// Execution
		startCycles := a.cpu.GetCycles()
		a.cpu.ExecuteInstruction()
		a.cycles += a.cpu.GetCycles() - startCycles
	}

	a.tickCards()
	a.executionTrace()
}

// drainCommands executes the commands queued by the frontend and returns what
// the run loop has to do about them
func (a *Apple2) drainCommands() commandOutcome {
	outcome := commandOutcomeNone
	for {
		select {
		case command := <-a.commandChannel:
			switch command.getId() {
			case CommandKill:
				return commandOutcomeKill
			case CommandPause:
				a.paused.Store(true)
			case CommandStart:
				if a.paused.Load() {
					a.paused.Store(false)
					outcome = commandOutcomeResync
				}
			case CommandPauseUnpause:
				a.paused.Store(!a.paused.Load())
				outcome = commandOutcomeResync
			default:
				// Execute the other commands
				a.executeCommand(command)
			}
		default:
			return outcome
		}
	}
}

func (a *Apple2) reset() {
	a.cpu.Reset()
	a.mmu.reset()
	for _, c := range a.cards {
		if c != nil {
			c.reset()
		}
	}
}

func (a *Apple2) tickCards() {
	for _, c := range a.tickerCards {
		c.tick()
	}
}

func (a *Apple2) executionTrace() {
	for _, v := range a.tracers {
		v.inspect()
	}
}

func (a *Apple2) dumpDebugInfo() {
	// See "Apple II Monitors Peeled"
	pageZeroSymbols := map[uint16]string{
		0x36: "CSWL",
		0x37: "CSWH",
		0x38: "KSWL",
		0x39: "KSWH",
		0xe2: "ACJVAFLDL", // Apple Pascal
		0xe3: "ACJVAFLDH", // Apple Pascal
		0xec: "JVBFOLDL",  // Apple Pascal
		0xed: "JVBFOLDH",  // Apple Pascal
		0xee: "JVAFOLDL",  // Apple Pascal
		0xef: "JVAFOLDH",  // Apple Pascal
	}
	fmt.Printf("Page zero values:\n")
	for _, k := range []uint16{0x36, 0x37, 0x38, 0x39, 0xe2, 0xe3, 0xec, 0xed, 0xee, 0xef} {
		d := a.mmu.physicalMainRAM.peek(k)
		fmt.Printf("  %v(0x%x): 0x%02x\n", pageZeroSymbols[k], k, d)
	}

	pc := uint16(0xc700)
	for pc < 0xc800 {
		line, newPc := a.cpu.DisasmInstruction(pc)
		fmt.Println(line)
		pc = newPc
	}
}
