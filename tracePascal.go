package izapple2

import "fmt"

type tracePascal struct {
	a           *Apple2
	skipConsole bool
}

const (
	pascalJvabfoldL uint16 = 0x00ec // Points to the BIOS entry points
	pascalJvabfoldH uint16 = 0x00ed // Points to the BIOS entry points
)

func newTracePascal() *tracePascal {
	var t tracePascal
	t.skipConsole = true
	return &t
}

func (t *tracePascal) connect(a *Apple2) {
	t.a = a
}

/*
See:

	https://archive.org/details/Hyde_P-Source-A_Guide_to_the_APPLE_Pascal_System_1983/page/n415/mode/1up?view=theater
	https://archive.org/details/Apple_II_Pascal_1.2_Device_and_Interrupt_Support_Tools_Manual

Experimental. Not sure the paramters for DREAD and DWRITE are correct.
*/
func (t *tracePascal) inspect() {
	if t.a.dmaActive {
		return
	}

	bios := uint16(t.a.mmu.physicalMainRAM.peek(pascalJvabfoldL)) +
		uint16(t.a.mmu.physicalMainRAM.peek(pascalJvabfoldH))<<8
	pc, _ := t.a.cpu.GetPCAndSP()
	if pc >= bios && pc < bios+0x5a {
		offset := uint8(pc - bios)

		if t.skipConsole && offset <= 0x03 {
			return
		}

		_, regA := t.a.cpu.GetCarryAndAcc()
		regAText := string(regA)
		if regA < 0x20 {
			regAText = "^" + string(regA+0x40)

		}

		fmt.Printf("Pascal BIOS call $%02x from $%04x ", offset, t.param(1))
		switch offset {
		case 0:
			fmt.Printf("CREAD()")
		case 3:
			fmt.Printf("CWRITE('%s'[%v])", regAText, regA)
		case 6:
			fmt.Printf("CINIT(BREAK=[$%04x], SYSCOM=[$%04x])",
				t.param(3), t.param(5))
		case 9:
			fmt.Printf("PWRITE('%s'[%v])", regAText, regA)
		case 12:
			fmt.Printf("PINIT()")
		case 15:
			fmt.Printf("DWRITE(UNIT=%v, BLOCK=%v, LEN=%v, DATA=[$%04x], DRIVE=%v, CONTROL=$%04x)",
				regA, t.param(3), t.param(5), t.param(7), t.param(9), t.param(11))
		case 18:
			fmt.Printf("DREAD(UNIT=%v, BLOCK=%v, LEN=%v, DATA=[$%04x], DRIVE=%v, CONTROL=$%04x)",
				regA, t.param(3), t.param(5), t.param(7), t.param(9), t.param(11))
		case 21:
			fmt.Printf("DINIT(DRIVE=%v)", regA)
		case 24:
			fmt.Printf("RREAD()")
		case 27:
			fmt.Printf("RWRITE('%s'[%v])", regAText, regA)
		case 30:
			fmt.Printf("RINIT()")
		case 33:
			fmt.Printf("GWRITE()")
		case 36:
			fmt.Printf("GRINIT()")
		case 39:
			fmt.Printf("PREAD()")
		case 42:
			fmt.Printf("CSTAT()")
		case 45:
			fmt.Printf("PSTAT()")
		case 48:
			fmt.Printf("DSTAT()")
		case 51:
			fmt.Printf("RSTAT()")
		case 54:
			fmt.Printf("CONCK()")
		case 57:
			fmt.Printf("UDRWI()")
		case 60:
			fmt.Printf("PSUBDRV()")

		default:
			fmt.Printf("<unknown>")
		}
		fmt.Printf("\n")
	}
}

/*
 See http://www.bitsavers.org/pdf/softech/softechPascalIV_intArch1981.pdf
 page 106
*/
//lint:ignore U1000 unused but stays as reference
func (t *tracePascal) inspectPerArchitectureGuide() {
	bios := uint16(t.a.mmu.physicalMainRAM.peek(pascalJvabfoldL)) +
		uint16(t.a.mmu.physicalMainRAM.peek(pascalJvabfoldH))<<8
	pc, _ := t.a.cpu.GetPCAndSP()
	if pc >= bios && pc < bios+0x5a {
		offset := uint8(pc - bios)

		if t.skipConsole && offset <= 0x03 {
			return
		}

		_, regA := t.a.cpu.GetCarryAndAcc()
		regAText := string(regA)
		if regA < 0x20 {
			regAText = "^" + string(regA+0x40)

		}

		fmt.Printf("Pascal BIOS call $%02x from $%04x ", offset, t.param(1))
		switch offset {
		// Console
		case 0x00:
			fmt.Printf("CONSOLEREAD()")
		case 0x03:
			fmt.Printf("CONSOLEWRITE('%s'[%v])", regAText, regA)
		case 0x06:
			fmt.Printf("CONSOLECTRL(BREAK=[%04x], SYSCOM=[%04x])",
				t.param(3), t.param(5))
		case 0x09:
			fmt.Printf("CONSOLESTAT(STATREC=[%04x], CONTROL=%04x)",
				t.param(3), t.param(5))

		// Printer
		case 0x0c:
			fmt.Printf("PRINTERREAD()")
		case 0x0f:
			fmt.Printf("PRINTERWRITE('%s'[%v])", regAText, regA)
		case 0x12:
			fmt.Printf("PRINTERCTRL()")
		case 0x15:
			fmt.Printf("PRINTERSTAT(STATREC=[%04x], CONTROL=%04x)",
				t.param(3), t.param(5))

		// Disk
		case 0x18:
			fmt.Printf("DISKREAD(BLOCK=%04x, LEN=%04x, DATA=[%04x], DRIVE=%v, CONTROL=%04x)",
				t.param(3), t.param(5), t.param(7), t.param(9), t.param(11))
		case 0x1b:
			fmt.Printf("DISKWRITE(BLOCK=%04x, LEN=%04x, DATA=[%04x], DRIVE=%v, CONTROL=%04x)",
				t.param(3), t.param(5), t.param(7), t.param(9), t.param(11))
		case 0x1e:
			fmt.Printf("DISKCTRL(DRIVE=%v)", regA)
		case 0x21:
			fmt.Printf("DISKSTAT(DRIVE=%v, STATREC=[%04x], CONTROL=%04x)",
				regA, t.param(3), t.param(5))

		// Remote
		case 0x24:
			fmt.Printf("REMOTEREAD()")
		case 0x27:
			fmt.Printf("REMOTEWRITE('%s'[%v])", regAText, regA)
		case 0x2a:
			fmt.Printf("REMOTECTRL()")
		case 0x2d:
			fmt.Printf("REMOTESTAT(STATREC=[%04x], CONTROL=%04x)",
				t.param(3), t.param(5))

		// User
		case 0x30:
			fmt.Printf("USERREAD(BLOCK=%04x, LEN=%04x, DATA=[%04x], DEVICE=%v, CONTROL=%04x)",
				t.param(3), t.param(5), t.param(7), t.param(9), t.param(11))
		case 0x33:
			fmt.Printf("USERWRITE(BLOCK=%04x, LEN=%04x, DATA=[%04x], DEVICE=%v, CONTROL=%04x)",
				t.param(3), t.param(5), t.param(7), t.param(9), t.param(11))
		case 0x36:
			fmt.Printf("USERCTRL(DEVICE=%v)", regA)
		case 0x39:
			fmt.Printf("USERSTAT(DEVICE=%v, STATREC=[%04x], CONTROL=%04x)",
				regA, t.param(3), t.param(5))

		// Sys
		case 0x3c:
			fmt.Printf("SYSREAD(BLOCK=%04x, LEN=%04x, DATA=[%04x], DEVICE=%v, CONTROL=%04x)",
				t.param(3), t.param(5), t.param(7), t.param(9), t.param(11))
		case 0x3f:
			fmt.Printf("SYSWRITE(BLOCK=%04x, LEN=%04x, DATA=[%04x], DEVICE=%v, CONTROL=%04x)",
				t.param(3), t.param(5), t.param(7), t.param(9), t.param(11))
		case 0x42:
			fmt.Printf("SYSCTRL(DEVICE=%v)", regA)
		case 0x43:
			fmt.Printf("SYSSTAT(DEVICE=%v, STATREC=[%04x], CONTROL=%04x)",
				regA, t.param(3), t.param(5))

		default:
			fmt.Printf("<unknown>")
		}
		fmt.Printf("\n")
	}
}

func (t *tracePascal) param(index uint8) uint16 {
	_, sp := t.a.cpu.GetPCAndSP()
	return uint16(t.a.mmu.Peek(0x100+uint16(sp+index))) +
		uint16(t.a.mmu.Peek(0x100+uint16(sp+index+1)))<<8
}
