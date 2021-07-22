package izapple2

import "fmt"

/*
  See:
    https://github.com/bobbimanners/Applecorn
	http://beebwiki.mdfs.net/Category:MOS_API
	http://beebwiki.mdfs.net/OSBYTEs
	http://mdfs.net/Docs/Comp/BBC/Osbyte00

*/

type traceApplecorn struct {
	a           *Apple2
	skipConsole bool
	osbyteNames [256]string
}

const (
	applecornMosVec uint16 = 0xffb9 // Start of the MOS entry points
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
	if pc >= applecornMosVec {
		regA, regX, regY := t.a.cpu.GetAXY()
		s := ""

		if !t.skipConsole {
			switch pc {
			case 0xffe0:
				s = fmt.Sprintf("OSNEWL()")
			case 0xffc8:
				ch := ""
				if regA >= 0x20 && regA < 0x7f {
					ch = string(regA)
				}
				s = fmt.Sprintf("OSNWRCH(A=%02x, '%v')", regA, ch)
			case 0xffcb:
				s = fmt.Sprintf("OSNRDCH()")
			case 0xffe7:
				s = fmt.Sprintf("OSRDCH()")
			case 0xffee:
				ch := ""
				if regA >= 0x20 && regA < 0x7f {
					ch = string(regA)
				}
				s = fmt.Sprintf("OSWRCH(A=%02x, '%v')", regA, ch)
			}

		}

		switch pc {
		case 0xffb9:
			s = "OSDRM(?)"
		case 0xffbf:
			s = "OSEVEN(?)"
		case 0xffc2:
			s = "OSINIT(?)"
		case 0xffc5:
			s = "OSREAD(?)"
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
		case 0xffe3:
			s = "OSASCI(?)"
		case 0xfff1:
			s = fmt.Sprintf("OSWORD(A=%02x,XY=%04x)", regA, uint16(regX)<<8+uint16(regY))
		case 0xfff4:
			s = fmt.Sprintf("OSBYTE('%s';A=%02x,X=%02x,Y=%02x)", t.osbyteNames[regA], regA, regX, regY)
		case 0xfff7:
			s = "OSCLI(?)"
		}

		if s != "" {
			fmt.Printf("BBC MOS call to $%04x %s\n", pc, s)
		}
	}
}
