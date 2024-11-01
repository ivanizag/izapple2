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

// Run starts the Apple2 emulation
func (a *Apple2) Run() {
	a.Start(false)
}

// Start the Apple2 emulation, can start paused
func (a *Apple2) Start(paused bool) {
	// Start the processor
	a.cpu.Reset()
	a.cycles = a.cpu.GetCycles()

	referenceTime := time.Now()
	speedReferenceTime := referenceTime
	speedReferenceCycles := uint64(0)

	a.paused = paused

	for {
		// Run cpu steps
		if !a.paused {
			if !a.dmaActive {
				// 6502 is running
				for i := 0; i < cpuSpinLoops && !a.dmaActive; i++ {
					// Conditional tracing
					// pc, _ := a.cpu.GetPCAndSP()
					// a.cpu.SetTrace(pc >= 0xc700 && pc < 0xc800)

					// Execution
					startCycles := a.cpu.GetCycles()
					a.cpu.ExecuteInstruction()
					a.cycles += a.cpu.GetCycles() - startCycles

					a.executionTrace()
				}
			} else {
				// a card, like the Z80 Softcard, is running
				card := a.cards[a.dmaSlot]
				for i := 0; i < cpuSpinLoops && a.dmaActive; i++ {
					card.runDMACycle()
					a.cycles++

					a.executionTrace()
				}
			}

			if a.cycleBreakpoint != 0 && a.cycles >= a.cycleBreakpoint {
				a.breakPoint = true
				a.cycleBreakpoint = 0
				a.paused = true
			}
		} else {
			time.Sleep(200 * time.Millisecond)
		}

		// Execute meta commands
		commandsPending := true
		for commandsPending {
			select {
			case command := <-a.commandChannel:
				switch command.getId() {
				case CommandKill:
					return
				case CommandPause:
					if !a.paused {
						a.paused = true
					}
				case CommandStart:
					if a.paused {
						a.paused = false
						referenceTime = time.Now()
						speedReferenceTime = referenceTime
					}
				case CommandPauseUnpause:
					a.paused = !a.paused
					referenceTime = time.Now()
					speedReferenceTime = referenceTime
				default:
					// Execute the other commands
					a.executeCommand(command)
				}
			default:
				commandsPending = false
			}
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

func (a *Apple2) reset() {
	a.cpu.Reset()
	a.mmu.reset()
	for _, c := range a.cards {
		if c != nil {
			c.reset()
		}
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
