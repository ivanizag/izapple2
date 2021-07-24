package izapple2

import (
	"fmt"
	"strings"
)

/*
  See:
    https://github.com/bobbimanners/Applecorn
	Chapter 45 https://raw.githubusercontent.com/bobbimanners/Applecorn/main/Manuals/Acorn%20BBC%20Micro%20User%20Guide.pdf
	http://beebwiki.mdfs.net/Category:MOS_API
	http://beebwiki.mdfs.net/OSBYTEs
	http://mdfs.net/Docs/Comp/BBC/Osbyte00

*/

type traceApplecorn struct {
	a           *Apple2
	skipConsole bool
	osbyteNames [256]string
	calls       []mosCallData
	lastDepth   int
}

type mosCallData struct {
	caller  uint16
	api     uint16
	a, x, y uint8
}

const (
	applecornMosVec   uint16 = 0xffb9 // Start of the MOS entry points
	applecornNoCaller uint16 = 0xffff
)

func newTraceApplecorn(a *Apple2, skipConsole bool) *traceApplecorn {
	var t traceApplecorn
	t.a = a
	t.skipConsole = skipConsole
	t.osbyteNames[0x7c] = "clear escape condition"
	t.osbyteNames[0x7d] = "set escape condition"
	t.osbyteNames[0x7e] = "ack detection of ESC"
	t.osbyteNames[0x81] = "Read key with time lim"
	t.osbyteNames[0x82] = "read high order address"
	t.osbyteNames[0x83] = "read bottom of user mem"
	t.osbyteNames[0x84] = "read top of user mem"
	t.osbyteNames[0x85] = "top user mem for mode"
	t.osbyteNames[0x86] = "read cursor pos"
	t.osbyteNames[0xDA] = "clear VDU queue"
	t.calls = make([]mosCallData, 0)
	return &t
}

func (t *traceApplecorn) inspect() {
	pc, sp := t.a.cpu.GetPCAndSP()
	if pc >= 0xd000 /*applecornMosVec*/ {
		regA, regX, regY := t.a.cpu.GetAXY()

		s := ""
		skip := false

		// Page 2 vectors
		switch pc {
		case t.vector(0x0208): // OSCLI vector
			pc = 0xfff7
		case t.vector(0x020A): // OSBYTE vector
			pc = 0xfff4
		case t.vector(0x020C): // OSWORD vector
			pc = 0xfff1
		case t.vector(0x020E): // OSWRCH vector
			pc = 0xffee
		case t.vector(0x0210): // OSRDCH vector
			pc = 0xffe0
		case t.vector(0x0212): // OSFILE vector
			pc = 0xffdd
		case t.vector(0x0214): // OSARGS vector
			pc = 0xffda
		case t.vector(0x0216): // OSBGET vector
			pc = 0xffd7
		case t.vector(0x0218): // OSBPUT vector
			pc = 0xffd4
		case t.vector(0x021A): // OSGBPB vector
			pc = 0xffd1
		case t.vector(0x021C): // OSFIND vector
			pc = 0xffce
		}

		// Jump area
		switch pc {
		case 0xffb9:
			s = "OSDRM(?)"
		case 0xffbc:
			s = "VDUCHR(?)"
		case 0xffbf:
			s = "OSEVEN(?)"
		case 0xffc2:
			s = "OSINIT(?)"
		case 0xffc5:
			s = "OSREAD(?)"
		case 0xffc8:
			ch := ""
			if regA >= 0x20 && regA < 0x7f {
				ch = string(regA)
			}
			s = fmt.Sprintf("OSNWRCH(A=%02x, '%v')", regA, ch)
			skip = t.skipConsole
		case 0xffcb:
			s = fmt.Sprintf("OSNRDCH()")
			skip = t.skipConsole
		case 0xffce:
			s = "OSFIND(?)"
		case 0xffd1:
			s = "OSGBPB(?)"
		case 0xffd4:
			s = "OSBPUT(?)"
		case 0xffd7:
			s = "OSBGET(?)"
		case 0xffda:
			s = "OSARGS(?)"
		case 0xffdd:
			s = "OSFILE(?)"
		case 0xffe0:
			s = fmt.Sprintf("OSRDCH()")
			skip = t.skipConsole
			/*
				case 0xffe3: // This fallbacks to OSWRCH
					s = "OSASCI(?)"
					skip = t.skipConsole
				case 0xffe7: // This fallbacks to OSWRCH
					s = fmt.Sprintf("OSNEWL()")
					skip = t.skipConsole
				case 0xffec: // This fallbacks to OSWRCH
					skip = t.skipConsole
					s = fmt.Sprintf("OSNECR()")
			*/
		case 0xffee:
			ch := ""
			if regA >= 0x20 && regA < 0x7f {
				ch = string(regA)
			}
			s = fmt.Sprintf("OSWRCH(A=%02x, '%v')", regA, ch)
			skip = t.skipConsole
		case 0xfff1:
			xy := uint16(regX) + uint16(regY)<<8
			switch regA {
			case 0: // Read line from input
				lineAddress := t.a.mmu.peekWord(xy)
				maxLength := t.a.mmu.Peek(xy + 2)
				s = fmt.Sprintf("OSWORD('read line';A=%02x,XY=%04x,BUF=%04x,MAX=%02x)", regA, xy, lineAddress, maxLength)
			default:
				s = fmt.Sprintf("OSWORD(A=%02x,XY=%04x)", regA, xy)
			}
		case 0xfff4:
			s = fmt.Sprintf("OSBYTE('%s';A=%02x,X=%02x,Y=%02x)", t.osbyteNames[regA], regA, regX, regY)
			//if regA == 0xda {
			//	t.a.cpu.Reset()
			//}
		case 0xfff7:
			s = "OSCLI(?)"
		}

		if !skip && s != "" {
			caller := t.a.mmu.peekWord(0x100+uint16(sp+1)) + 1
			t.calls = append(t.calls, mosCallData{caller, pc, regA, regX, regY})
			if len(t.calls) > t.lastDepth {
				// Reentrant call, first of block
				fmt.Println()
			}
			if len(t.calls) > 1 {
				// Reentrant call
				fmt.Printf("%s", strings.Repeat("  ", len(t.calls)))
			}
			fmt.Printf("BBC MOS call to $%04x %s ", pc, s)
			t.lastDepth = len(t.calls)
		}
	}

	if len(t.calls) > 0 && pc == t.calls[len(t.calls)-1].caller {
		// Returning from the call
		regA, regX, regY := t.a.cpu.GetAXY()
		call := t.calls[len(t.calls)-1]
		s := ""
		switch call.api {
		case 0xfff1: // OSWORD
			cbAddress := uint16(call.x) + uint16(call.y)<<8
			switch call.a {
			case 0: // Read line from input
				lineAddress := t.a.mmu.peekWord(cbAddress)
				line := t.getString(lineAddress, regY)
				s = fmt.Sprintf(",line='%s'", line)
			}
		}

		fmt.Printf("=> (A=%02x,X=%02x,Y=%02x%s)\n", regA, regX, regY, s)
		t.calls = t.calls[:len(t.calls)-1]
	}
}

func (t *traceApplecorn) getString(address uint16, length uint8) string {
	s := ""
	for i := uint8(0); i < length; i++ {
		ch := t.a.mmu.Peek(address + uint16(i))
		s = s + string(ch)
	}
	return s
}

func (t *traceApplecorn) vector(address uint16) uint16 {
	return t.a.mmu.peekWord(address)
}
