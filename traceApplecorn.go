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
	wasInKernel bool
}

type mosCallData struct {
	api     uint16
	a, x, y uint8
	skipLog bool
}

const (
	applecornKernelStart    uint16 = 0xc000 // Code above this is out of BBC territory
	applecornTraceAreaStart uint16 = 0xd000
	applecornNoCaller       uint16 = 0xffff
	applecornRomTitle       uint16 = 0x8009
)

func newTraceApplecorn(skipConsole bool) *traceApplecorn {
	var t traceApplecorn
	t.skipConsole = skipConsole
	t.osbyteNames[0x02] = "select input device"
	t.osbyteNames[0x03] = "select output device"
	t.osbyteNames[0x7c] = "clear escape condition"
	t.osbyteNames[0x7d] = "set escape condition"
	t.osbyteNames[0x7e] = "ack detection of ESC"
	t.osbyteNames[0x7f] = "check for end-of-file on an opened file"
	t.osbyteNames[0x80] = "read ADC channel"
	t.osbyteNames[0x81] = "read key with time lim"
	t.osbyteNames[0x82] = "read high order address"
	t.osbyteNames[0x83] = "read bottom of user mem"
	t.osbyteNames[0x84] = "read top of user mem"
	t.osbyteNames[0x85] = "top user mem for mode"
	t.osbyteNames[0x86] = "read cursor pos"
	t.osbyteNames[0x8b] = "set filing system options"
	t.osbyteNames[0xDA] = "clear VDU queue"
	return &t
}

func (t *traceApplecorn) connect(a *Apple2) {
	t.a = a
}

func (t *traceApplecorn) inspect() {
	if !t.a.mmu.altMainRAMActiveRead {
		// We want to trace only the activity on the Acorn memory space
		return
	}

	pc, sp := t.a.cpu.GetPCAndSP()

	if pc == 0x8000 {
		activeROM := t.getTerminatedString(applecornRomTitle, 0)
		regA, _, _, _ := t.a.cpu.GetAXYP()
		fmt.Printf("BBC MOS call to $%04x LANGUAGE(ROM=\"%s\", A=%02x)\n", pc, activeROM, regA)
	} else if pc == 0x8003 {
		activeROM := t.getTerminatedString(applecornRomTitle, 0)
		service, _, regY, _ := t.a.cpu.GetAXYP()
		switch service {
		case 4: // OSCLI
			address := t.a.mmu.peekWord(0xf2 + uint16(regY))
			command := t.getTerminatedString(address, 0x0d)
			fmt.Printf("BBC MOS call to $%04x SERVICE_OSCLI(ROM=\"%s\", A=%v, \"%s\")\n",
				pc, activeROM, service, command)
		case 6: // Error
			address := t.a.mmu.peekWord(0xfd)
			faultNumber := t.a.mmu.Peek(address)
			faultMessage := address + 1
			faultString := t.getTerminatedString(faultMessage, 0)
			fmt.Printf("BBC MOS call to $%04x SERVICE_ERROR(ROM=\"%s\", A=%v, #=%v, \"%s\")\n",
				pc, activeROM, service, faultNumber, faultString)
		case 7: // OSBYTE
			pA := t.a.mmu.Peek(0xef)
			pX := t.a.mmu.Peek(0xf0)
			pY := t.a.mmu.Peek(0xf1)
			fmt.Printf("BBC MOS call to $%04x SERVICE_OSBYTE%02x(ROM=\"%s\", A=%v, pX=%02x, pY=%02x)\n",
				pc, pA, activeROM, service, pX, pY)
		case 8: // OSWORD
			pA := t.a.mmu.Peek(0xef)
			fmt.Printf("BBC MOS call to $%04x SERVICE_OSWORD%02x(ROM=\"%s\", A=%v)\n",
				pc, pA, activeROM, service)
		case 9: // *HELP
			address := t.a.mmu.peekWord(0xf2 + uint16(regY))
			command := t.getTerminatedString(address, 0x0d)
			fmt.Printf("BBC MOS call to $%04x SERVICE_HELP(ROM=\"%s\", A=%v, \"%s\")\n",
				pc, activeROM, service, command)
		default:
			fmt.Printf("BBC MOS call to $%04x SERVICE(ROM=\"%s\", A=%v)\n",
				pc, activeROM, service)

		}
	}

	inKernel := pc >= applecornKernelStart
	if !t.wasInKernel && inKernel && pc >= applecornTraceAreaStart {
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
		case t.vector(0xfffe): // BRK vector
			pc = 0xfffe
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
			s = "OSNRDCH()"
			skip = t.skipConsole
		case 0xffce:
			if regA == 0 {
				s = fmt.Sprintf("OSFIND('close',HANDLE=%v", regY)
			} else {
				filenameAddress := uint16(regX) + uint16(regY)<<8
				filename := t.getTerminatedString(filenameAddress, 0x0d)
				s = fmt.Sprintf("OSFIND('open',FILE='%s')", filename)
			}
		case 0xffd1:
			s = "OSGBPB(?)"
		case 0xffd4:
			s = "OSBPUT(?)"
		case 0xffd7:
			s = "OSBGET(?)"
		case 0xffda:
			s = "OSARGS(?)"
			s = fmt.Sprintf("OSARGS(HANDLE=%v,A=%02x)", regY, regA)
		case 0xffdd:
			controlBlock := uint16(regX) + uint16(regY)<<8
			filenameAddress := t.a.mmu.peekWord(controlBlock)
			filename := t.getTerminatedString(filenameAddress, 0x0d)
			s = fmt.Sprintf("OSFILE(A=%02x,FILE='%s')", regA, filename)
		case 0xffe0:
			s = "OSRDCH()"
			skip = t.skipConsole
		case 0xffe3: // This fallbacks to OSWRCH
			ch := ""
			if regA >= 0x20 && regA < 0x7f {
				ch = string(regA)
			}
			s = fmt.Sprintf("OSASCI(A=%02x, '%v')", regA, ch)
			skip = t.skipConsole
		case 0xffe7: // This fallbacks to OSWRCH
			s = "OSNEWL()"
			skip = t.skipConsole
		case 0xffec: // This fallbacks to OSWRCH
			skip = t.skipConsole
			s = "OSNECR()"
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
			xy := uint16(regX) + uint16(regY)<<8
			command := t.getTerminatedString(xy, 0x0d)
			s = fmt.Sprintf("OSCLI(\"%s\")", command)
		case 0xfffe:
			address := t.a.mmu.peekWord(0x100+uint16(sp+2)) - 1
			faultNumber := t.a.mmu.Peek(address)
			faultMessage := address + 1
			faultString := t.getTerminatedString(faultMessage, 0)
			s = fmt.Sprintf("BRK(number=%v,message='%s')", faultNumber, faultString)
		}

		if s == "UNKNOWN" && t.skipConsole {
			// Let's also skip not known calls
			skip = true
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
				line := t.getTerminatedString(lineAddress, '\r')
				s = fmt.Sprintf(",line='%s',cbaddress=%04x,lineAddress=%04x", line, cbAddress, lineAddress)
			}
		}

		fmt.Printf("=> (A=%02x,X=%02x,Y=%02x%s)\n", regA, regX, regY, s)
	}

	t.wasInKernel = inKernel
}

func (t *traceApplecorn) getTerminatedString(address uint16, terminator uint8) string {
	s := ""
	for {
		ch := t.a.mmu.Peek(address)
		if ch == terminator {
			break
		}
		s += string(ch)
		address++
	}
	return s
}

func (t *traceApplecorn) vector(address uint16) uint16 {
	return t.a.mmu.peekWord(address)
}
