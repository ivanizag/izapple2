package izapple2

import (
	"sync/atomic"

	"github.com/ivanizag/iz6502"
	"github.com/ivanizag/izapple2/screen"
)

// Apple2 represents all the components and state of the emulated machine
type Apple2 struct {
	Name    string
	cpu     *iz6502.State
	mmu     *memoryManager
	io      *ioC0Page
	video   screen.VideoSource
	cg      *CharacterGenerator
	cards   [8]Card
	tracers []executionTracer

	softVideoSwitch *SoftVideoSwitch
	board           string
	isApple2e       bool
	hasLowerCase    bool
	isFourColors    bool // An Apple II without the 6 color mod
	commandChannel  chan command

	dmaActive bool
	dmaSlot   int

	cycles               uint64
	cycleDurationNs      float64 // Current speed. Inverse of the cpu clock in Ghz
	fastRequestsCounter  int32
	cycleBreakpoint      uint64
	breakPoint           bool
	profile              bool
	showSpeed            bool
	paused               bool
	cpuTrace             bool
	forceCaps            bool
	removableMediaDrives []drive
}

// GetCards returns the array of inserted cards
func (a *Apple2) GetCards() [8]Card {
	return a.cards
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

// IsPaused returns true when emulator is paused
func (a *Apple2) IsPaused() bool {
	return a.paused
}

func (a *Apple2) GetCycles() uint64 {
	return a.cycles
}

// SetCycleBreakpoint sets a cycle number to pause the emulator. 0 to disable
func (a *Apple2) SetCycleBreakpoint(cycle uint64) {
	a.cycleBreakpoint = cycle
	a.breakPoint = false
}

func (a *Apple2) BreakPoint() bool {
	return a.breakPoint
}

// IsProfiling returns true when profiling
func (a *Apple2) IsProfiling() bool {
	return a.profile
}

// IsForceCaps returns true when all letters are forced to upper case
func (a *Apple2) IsForceCaps() bool {
	return a.forceCaps
}

func (a *Apple2) GetCgPageInfo() (int, int) {
	return a.cg.getPage(), a.cg.getPages()
}

func (a *Apple2) RequestFastMode() {
	// Note: if the fastMode is shorter than maxWaitDuration, there won't be any gain.
	atomic.AddInt32(&a.fastRequestsCounter, 1)
}

func (a *Apple2) ReleaseFastMode() {
	atomic.AddInt32(&a.fastRequestsCounter, -1)
}

func (a *Apple2) registerRemovableMediaDrive(d drive) {
	a.removableMediaDrives = append(a.removableMediaDrives, d)
}

func (a *Apple2) GetVideoSource() screen.VideoSource {
	return a.video
}
