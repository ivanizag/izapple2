package izapple2

import (
	"github.com/ivanizag/izapple2/screen"
)

/*
What a program driving the machine with RunCycles needs to look at it and to
change its media, the way a person at the keyboard would: the text on the
screen, the memory, and the diskettes. They must be called from the goroutine
that calls RunCycles, or while the machine is not running.
*/

// ScreenText returns the text shown on the screen, in 40 or 80 columns or on
// a Videx card in slot 3, with a line per row. Inverse and flashing
// characters are returned as the normal ones.
func (a *Apple2) ScreenText() string {
	switch videx := a.cards[3].(type) {
	case *CardVidexVideoterm:
		return videx.getText()
	case *CardVidexUltraterm:
		return videx.getText()
	}

	is80Columns := a.io.isSoftSwitchActive(ioFlag80Col)
	isSecondPage := a.io.isSoftSwitchActive(ioFlagSecondPage) && !a.mmu.store80Active
	isAltText := a.io.isSoftSwitchActive(ioFlagAltChar)
	return screen.RenderTextModeString(a.video, is80Columns, isSecondPage, isAltText, a.hasLowerCase, false)
}

// Peek returns the byte the processor would read at an address, with the
// memory banks as they are now. The I/O area from 0xc000 to 0xcfff returns
// 0, as reading it can change the state of the machine.
func (a *Apple2) Peek(address uint16) uint8 {
	if address >= 0xc000 && address < 0xd000 {
		return 0
	}
	return a.mmu.Peek(address)
}

// GetPC returns the program counter of the processor
func (a *Apple2) GetPC() uint16 {
	pc, _ := a.cpu.GetPCAndSP()
	return pc
}

// LoadDisk inserts a file on a removable media drive, numbered in the order
// of GetRemovableMediaDrives, right away. SendLoadDisk does the same from
// another goroutine, and only reports the errors on the console.
func (a *Apple2) LoadDisk(unit int, path string) error {
	return a.changeDisk(unit, path)
}
