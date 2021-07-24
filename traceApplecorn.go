package izapple2

import (
	"fmt"
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
	call        mosCallData
	lastDepth   int
	wasInKernel bool
}

type mosCallData struct {
	api     uint16
	a, x, y uint8
	skipLog bool
}

const (
	applecornKernelStart uint16 = 0xc000 // Code above this is out of BBC territory
	applecornNoCaller    uint16 = 0xffff
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
	return &t
}

func (t *traceApplecorn) inspect() {
	pc, _ := t.a.cpu.GetPCAndSP()
	inKernel := pc >= applecornKernelStart

	if !t.wasInKernel && inKernel {
		regA, regX, regY, _ := t.a.cpu.GetAXYP()

		s := "UNKNOWN"
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
		case 0xffe3: // This fallbacks to OSWRCH
			s = "OSASCI(?)"
			skip = t.skipConsole
		case 0xffe7: // This fallbacks to OSWRCH
			s = fmt.Sprintf("OSNEWL()")
			skip = t.skipConsole
		case 0xffec: // This fallbacks to OSWRCH
			skip = t.skipConsole
			s = fmt.Sprintf("OSNECR()")
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
		case 0xfff7:
			s = "OSCLI(?)"
		}

		t.call.api = pc
		t.call.a = regA
		t.call.x = regX
		t.call.y = regY
		t.call.skipLog = skip
		if !skip {
			fmt.Printf("BBC MOS call to $%04x %s ", pc, s)
		}
	}

	if t.wasInKernel && !inKernel && !t.call.skipLog {
		// Returning from the call
		regA, regX, regY, _ := t.a.cpu.GetAXYP()
		s := ""
		switch t.call.api {
		case 0xfff1: // OSWORD
			cbAddress := uint16(t.call.x) + uint16(t.call.y)<<8
			switch t.call.a {
			case 0: // Read line from input
				lineAddress := t.a.mmu.peekWord(cbAddress)
				line := t.getString(lineAddress, regY)
				s = fmt.Sprintf(",line='%s'", line)

				t.a.cpu.SetTrace(true)

			}
		}

		fmt.Printf("=> (A=%02x,X=%02x,Y=%02x%s)\n", regA, regX, regY, s)
	}

	t.wasInKernel = inKernel
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
