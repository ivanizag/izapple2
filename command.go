package izapple2

import "fmt"

const (
	// CommandToggleSpeed toggles cpu speed between full speed and actual Apple II speed
	CommandToggleSpeed = iota + 1
	// CommandShowSpeed toggles printinf the current freq in Mhz
	CommandShowSpeed
	// CommandDumpDebugInfo dumps useful info
	CommandDumpDebugInfo
	// CommandNextCharGenPage cycles the CharGen page if several
	CommandNextCharGenPage
	// CommandToggleCPUTrace toggle tracing of CPU execution
	CommandToggleCPUTrace
	// CommandKill stops the cpu execution loop
	CommandKill
	// CommandReset executes a 6502 reset
	CommandReset
	// CommandPauseUnpause allows the Pause button to freeze the emulator for a coffee break
	CommandPauseUnpause
	// CommandPause pauses the emulator
	CommandPause
	// CommandStart restarts the emulator
	CommandStart
	// CommandToggleIIc80ColumnSwitch moves the 80/40 switch of the Apple IIc
	CommandToggleIIc80ColumnSwitch
	// CommandComplex for commands that use a struct with parameters
	CommandComplex
)

type command interface {
	getId() int
}

type commandSimple struct {
	id int
}

type commandLoadDisk struct {
	drive int
	path  string
}

type commandLoadTape struct {
	path string
}

func (c *commandSimple) getId() int {
	return c.id
}

func (c *commandLoadDisk) getId() int {
	return CommandComplex
}

func (c *commandLoadTape) getId() int {
	return CommandComplex
}

func (a *Apple2) queueCommand(c command) {
	a.commandChannel <- c
}

// SendCommand enqueues a command to the emulator thread
func (a *Apple2) SendCommand(commandId int) {
	var c commandSimple
	c.id = commandId
	a.queueCommand(&c)
}

func (a *Apple2) SendLoadDisk(drive int, path string) {
	var c commandLoadDisk
	c.drive = drive
	c.path = path
	a.queueCommand(&c)
}

// SendLoadTape puts a tape in the cassette recorder from another goroutine,
// as LoadTape, reporting the errors on the console
func (a *Apple2) SendLoadTape(path string) {
	a.queueCommand(&commandLoadTape{path: path})
}

func (a *Apple2) executeCommand(command command) {
	switch command.getId() {
	case CommandToggleSpeed:
		if a.cycleDurationNs == 0 {
			// fmt.Println("Slow")
			a.cycleDurationNs = 1000.0 / CPUClockMhz
		} else {
			// fmt.Println("Fast")
			a.cycleDurationNs = 0
		}
	case CommandShowSpeed:
		fmt.Printf("Freq: %f Mhz\n", a.GetCurrentFreqMHz())
	case CommandDumpDebugInfo:
		a.dumpDebugInfo()
	case CommandNextCharGenPage:
		a.cg.nextPage()
		fmt.Printf("Chargen page %v\n", a.cg.page)
	case CommandToggleCPUTrace:
		a.cpuTrace = !a.cpuTrace
		a.cpu.SetTrace(a.cpuTrace)
	case CommandReset:
		a.reset()
	case CommandToggleIIc80ColumnSwitch:
		if !a.isApple2c {
			fmt.Println("Only the Apple IIc has the 80/40 switch")
		} else {
			a.SetIIc80ColumnSwitch(!a.iic80ColumnSwitch)
			if a.iic80ColumnSwitch {
				fmt.Println("80/40 switch up, on 80 columns")
			} else {
				fmt.Println("80/40 switch down, on 40 columns")
			}
		}
	case CommandComplex:
		switch t := command.(type) {
		case *commandLoadDisk:
			err := a.changeDisk(t.drive, t.path)
			if err != nil {
				fmt.Printf("Could no load file %v\n%v\n", t.path, err)
			}
		case *commandLoadTape:
			err := a.LoadTape(t.path)
			if err != nil {
				fmt.Printf("Could no load the tape %v\n%v\n", t.path, err)
			}
		}
	}
}

func (a *Apple2) changeDisk(unit int, path string) error {
	if unit >= 0 && unit < len(a.removableMediaDrives) {
		return a.removableMediaDrives[unit].drive.insertDiskette(path)
	}
	return fmt.Errorf("unit %v not defined", unit)
}
