package izapple2

import (
	"fmt"
)

/*
  See:
	https://github.com/davidgiven/cpm65
*/

type traceCpm65 struct {
	a           *Apple2
	skipConsole bool
}

const (
	cpm65BdosEntrypoint uint16 = 0xeb8f // TODO: find the correct address dinamically
	cpm65BiosEntrypoint uint16 = 0xfdce // TODO: find the correct address dinamically
)

func newTraceCpm65(skipConsole bool) *traceCpm65 {
	var t traceCpm65
	t.skipConsole = skipConsole
	return &t
}

func (t *traceCpm65) connect(a *Apple2) {
	t.a = a
}

func (t *traceCpm65) inspect() {
	if t.a.dmaActive {
		return
	}

	pc, _ := t.a.cpu.GetPCAndSP()
	if pc == cpm65BdosEntrypoint {
		regA, regX, regY, _ := t.a.cpu.GetAXYP()
		param := uint16(regX)<<8 | uint16(regA)
		switch regY {
		case 2: // CONSOLE_OUTPUT
			if !t.skipConsole {
				fmt.Printf("CPM65 BDOS call %02v:%s with \"%c\"\n", regY, bdos65CodeToName(regY), regA)
			}
		case 9: // WRITE_STRING
			if !t.skipConsole {
				text := t.getCpm65String(param)
				fmt.Printf("CPM65 BDOS call %02v:%s with \"%s\"\n", regY, bdos65CodeToName(regY), text)
			}
		default:
			fmt.Printf("CPM65 BDOS call %02v:%s($%04x)\n", regY, bdos65CodeToName(regY), param)
		}
	}

	if pc == cpm65BiosEntrypoint {
		regA, regX, regY, _ := t.a.cpu.GetAXYP()
		param := uint16(regX)<<8 | uint16(regA)
		if regY > 2 /*CONOUT*/ || !t.skipConsole {
			switch regY {
			default:
				fmt.Printf("CPM65 BIOS call %02v:%s($%04x)\n", regY, bios65CodeToName(regY), param)
			}
		}
	}
}

var cpm65BdosNames = []string{
	"EXIT_PROGRAM",           // 0
	"CONSOLE_INPUT",          // 1
	"CONSOLE_OUTPUT",         // 2
	"AUX_INPUT",              // 3
	"AUX_OUTPUT",             // 4
	"PRINTER_OUTPUT",         // 5
	"DIRECT_IO",              // 6
	"GET_IO_BYTE",            // 7
	"SET_IO_BYTE",            // 8
	"WRITE_STRING",           // 9
	"READ_LINE",              // 10
	"CONSOLE_STATUS",         // 11
	"GET_VERSION",            // 12
	"RESET_DISKS",            // 13
	"SELECT_DISK",            // 14
	"OPEN_FILE",              // 15
	"CLOSE_FILE",             // 16
	"FIND_FIRST",             // 17
	"FIND_NEXT",              // 18
	"DELETE_FILE",            // 19
	"READ_SEQUENTIAL",        // 20
	"WRITE_SEQUENTIAL",       // 21
	"CREATE_FILE",            // 22
	"RENAME_FILE",            // 23
	"GET_LOGIN_BITMAP",       // 24
	"GET_CURRENT_DRIVE",      // 25
	"SET_DMA_ADDRESS",        // 26
	"GET_ALLOCATION_BITMAP",  // 27
	"SET_DRIVE_READONLY",     // 28
	"GET_READONLY_BITMAP",    // 29
	"SET_FILE_ATTRIBUTES",    // 30
	"GET_DPB",                // 31
	"GET_SET_USER_NUMBER",    // 32
	"READ_RANDOM",            // 33
	"WRITE_RANDOM",           // 34
	"COMPUTE_FILE_SIZE",      // 35
	"COMPUTE_RANDOM_POINTER", // 36
	"RESET_DISK",             // 37
	"GET_BIOS",               // 38
	"",                       // 39
	"WRITE_RANDOM_FILLED",    // 40
	"GETZP",                  // 41
	"GETTPA",                 // 42
	"PARSE_FILENAME",         // 43
}

var cpm65BiosNames = []string{
	"CONST",    // 0
	"CONIN",    // 1
	"CONOUT",   // 2
	"SELDSK",   // 3
	"SETSEC",   // 4
	"SETDMA",   // 5
	"READ",     // 6
	"WRITE",    // 7
	"RELOCATE", // 8
	"GETTPA",   // 9
	"SETTPA",   // 10
	"GETZP",    // 11
	"SETZP",    // 12
	"SETBANK",  // 13
	"ADDDRV",   // 14
	"FINDDRV",  // 15
}

func bdos65CodeToName(code uint8) string {
	if code < uint8(len(cpm65BdosNames)) {
		return cpm65BdosNames[code]
	}
	return fmt.Sprintf("BDOS_%d", code)
}

func bios65CodeToName(code uint8) string {
	if code < uint8(len(cpm65BiosNames)) {
		return cpm65BiosNames[code]
	}
	return fmt.Sprintf("BIOS_%d", code)
}

func (t *traceCpm65) getCpm65String(address uint16) string {
	s := ""
	for {
		ch := t.a.mmu.Peek(address)
		if ch == '$' || ch == 0 {
			break
		}
		s += string(ch)
		address++
	}
	return s
}
