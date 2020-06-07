package apple2

import "fmt"

type traceProDOS struct {
	a              *Apple2
	callPending    bool // We assume MLI is not reentrant
	functionCode   uint8
	paramsAdddress uint16
	returnAddress  uint16
}

const (
	mliAddress uint16 = 0xbf00
	biAddress  uint16 = 0xbe03
)

func newTraceProDOS(a *Apple2) *traceProDOS {
	var t traceProDOS
	t.a = a
	return &t
}

func (t *traceProDOS) inspect() {
	pc, ps := t.a.cpu.GetPCAndSP()
	if pc == mliAddress {
		/*
			MLI has been called (provided we are running proDOS and the proper page)
			Calls to MLI must be:
			   JSR $BF00
			   DFB function_code
			   DW  addr_op_parms
		*/
		if t.callPending {
			if t.functionCode == 0x65 {
				// QUIT when successfull does not return
				fmt.Printf("Ok \n")
			} else {
				fmt.Print("<there was a call pending>\n")
			}
		}
		caller := uint16(t.a.mmu.Peek(0x100+uint16(ps+1))) +
			uint16(t.a.mmu.Peek(0x100+uint16(ps+2)))<<8 - 2
		t.functionCode = t.a.mmu.Peek(caller + 3)
		t.paramsAdddress = uint16(t.a.mmu.Peek(caller+4)) + uint16(t.a.mmu.Peek(caller+5))<<8
		t.returnAddress = caller + 6
		fmt.Printf("MLI call $%02x from $%04x", t.functionCode, caller)
		switch t.functionCode {
		case 0x40:
			fmt.Printf(" ALLOC_INTERRUPT()")
		case 0x41:
			fmt.Printf(" DEALLOC_INTERRUPT()")
		case 0x65:
			fmt.Printf(" QUIT()")
		case 0x80:
			fmt.Printf(" READ_BLOCK(unit=%s, block=$%04x)", parseUnit(t.paramByte(1)), t.paramWord(4))
		case 0x81:
			fmt.Printf(" WRITE_BLOCK(unit=%s, block=$%04x)", parseUnit(t.paramByte(1)), t.paramWord(4))
		case 0x82:
			fmt.Printf(" GET_TIME()")
		case 0xc0:
			fmt.Printf(" CREATE(\"%s\")", t.paramString(1))
		case 0xc1:
			fmt.Printf(" DESTROY(\"%s\")", t.paramString(1))
		case 0xc2:
			fmt.Printf(" RENAME(old=\"%s\", new=\"%s\")", t.paramString(1), t.paramString(3))
		case 0xc3:
			fmt.Printf(" GET_FILE_INFO(\"%s\")", t.paramString(1))
		case 0xc4:
			fmt.Printf(" SET_FILE_INFO(\"%s\")", t.paramString(1))
		case 0xc5:
			fmt.Printf(" ONLINE(unit=%s)", parseUnit(t.paramByte(1)))
		case 0xc6:
			fmt.Printf(" SET_PREFIX(\"%s\")", t.paramString(1))
		case 0xc7:
			fmt.Printf(" GET_PREFIX()")
		case 0xc8:
			fmt.Printf(" OPEN(\"%s\")", t.paramString(1))
		case 0xc9:
			fmt.Printf(" NEWLINE(ref=%v, mask=$%02x, char=$%02x)", t.paramByte(1), t.paramByte(2), t.paramByte(3))
		case 0xca:
			fmt.Printf(" READ(ref=%v, len=%v)", t.paramByte(1), t.paramWord(4))
		case 0xcb:
			fmt.Printf(" WRITE(ref=%v, len=%v)", t.paramByte(1), t.paramWord(4))
		case 0xcc:
			fmt.Printf(" CLOSE(ref=%v)", t.paramByte(1))
		case 0xcd:
			fmt.Printf(" FLUSH(ref=%v)", t.paramByte(1))
		case 0xce:
			fmt.Printf(" SET_MARK(ref=%v, pos=%v)", t.paramByte(1), t.paramLen(2))
		case 0xcf:
			fmt.Printf(" GET_MARK(ref=%v)", t.paramByte(1))
		case 0xd1:
			fmt.Printf(" GET_EOF(ref=%v)", t.paramByte(1))
		case 0xd2:
			fmt.Printf(" SET_BUF(ref=%v)", t.paramByte(1))
		case 0xd3:
			fmt.Printf(" GET_BUF(ref=%v)", t.paramByte(1))
		}
		fmt.Printf(" => ")

		t.callPending = true
	} else if t.callPending && pc == t.returnAddress {
		error, acc := t.a.cpu.GetCarryAndAcc()
		if error {
			fmt.Printf("error $%02x: %v\n", acc, getMliErrorText(acc))
		} else {
			switch t.functionCode {
			case 0x82: // Get Time
				// Globals will be updated
				date := uint16(t.a.mmu.Peek(0xbf90)) + uint16(t.a.mmu.Peek(0xbf91))<<8
				minute := t.a.mmu.Peek(0xbf92)
				hour := t.a.mmu.Peek(0xbf93)
				fmt.Printf("%04v-%02v-%02v %02v:%02v\n",
					date>>9+1900, (date>>5)&0x1f, date&0x1f, // Review Y2K
					hour, minute)
			case 0xc5: // Online
				dataAddress := t.paramWord(2)
				for {
					b := t.a.mmu.Peek(dataAddress)
					dataAddress++
					if b == 0 {
						fmt.Printf("\n")
						break
					}
					unit := parseUnit(b)
					size := b & 0xf
					if size != 0 {
						// No error
						name := ""
						for i := uint8(0); i < size; i++ {
							name += string(t.a.mmu.Peek(dataAddress+uint16(i)) & 0x7f)
						}
						fmt.Printf("%s: \"%s\" ", unit, name)
					} else {
						err := t.a.mmu.Peek(dataAddress)
						fmt.Printf("%s: error $%02x ", unit, err)
					}
					if t.paramByte(1) != 0 {
						fmt.Printf("\n")
						break // Only one entry requested
					}
					dataAddress += 15
				}
			case 0xc7: // Get prefix
				fmt.Printf("\"%v\"\n", t.paramString(1))
			case 0xc8: // Open file
				fmt.Printf("ref: %v\n", t.paramByte(5))
			case 0xca: // Read
				fmt.Printf("%v bytes read \n", t.paramByte(6))
			case 0xcb: // Write
				fmt.Printf("%v bytes written \n", t.paramByte(6))
			case 0xcf: // File position
				fmt.Printf("%v\n", t.paramLen(2))
			case 0xd1: // File size
				fmt.Printf("%v bytes\n", t.paramLen(2))
			default:
				fmt.Printf("Ok\n")
			}
		}
		t.callPending = false
	} else if pc == biAddress {
		s := ""
		for i := uint16(1); i < 256; i++ {
			ch := t.a.mmu.Peek(0x200 + i)
			if ch == 0 || ch == 0x8d {
				break
			}
			s += string(ch)
		}
		fmt.Printf("Prodos BI exec: \"%s\".\n", s)
	}
}

func (t *traceProDOS) paramByte(pos uint16) uint8 {
	return t.a.mmu.Peek(t.paramsAdddress + pos)
}

func (t *traceProDOS) paramWord(pos uint16) uint16 {
	// Two bytes
	return uint16(t.a.mmu.Peek(t.paramsAdddress+pos)) + uint16(t.a.mmu.Peek(t.paramsAdddress+pos+1))<<8
}

func (t *traceProDOS) paramLen(pos uint16) uint32 {
	// Three bytes
	return uint32(t.paramWord(pos)) + uint32(t.paramByte(pos+2))<<16
}

func (t *traceProDOS) paramString(pos uint16) string {
	address := t.paramWord(pos)
	size := t.a.mmu.Peek(address)
	s := ""
	for i := uint8(0); i < size; i++ {
		s += string(t.a.mmu.Peek(address+1+uint16(i)) & 0x7f)
	}
	return s
}

func parseUnit(unit uint8) string {
	if unit == 0 {
		return "All"
	}
	drive := unit >> 7
	slot := (unit >> 4) & 7
	return fmt.Sprintf("S%v,D%v", slot, drive+1)
}

func getMliErrorText(code uint8) string {
	// From https://prodos8.com/docs/techref/quick-reference-card/
	switch code {
	case 0x00:
		return "No error"
	case 0x01:
		return "Bad system call number"
	case 0x04:
		return "Bad system call parameter count"
	case 0x25:
		return "Interrupt table full"
	case 0x27:
		return "I/O error"
	case 0x28:
		return "No device connected"
	case 0x2B:
		return "Disk write protected"
	case 0x2E:
		return "Disk switched"
	case 0x40:
		return "Invalid pathname"
	case 0x42:
		return "Maximum number of files open"
	case 0x43:
		return "Invalid reference number"
	case 0x44:
		return "Directory not found"
	case 0x45:
		return "Volume not found"
	case 0x46:
		return "File not found"
	case 0x47:
		return "Duplicate filename"
	case 0x48:
		return "Volume full"
	case 0x49:
		return "Volume directory full"
	case 0x4A:
		return "Incompatible file format, also a ProDOS directory"
	case 0x4B:
		return "Unsupported storage_type"
	case 0x4C:
		return "End of file encountered"
	case 0x4D:
		return "Position out of range"
	case 0x4E:
		return "File access error, also file locked"
	case 0x50:
		return "File is open"
	case 0x51:
		return "Directory structure damaged"
	case 0x52:
		return "Not a ProDOS volume"
	case 0x53:
		return "Invalid system call parameter"
	case 0x55:
		return "Volume Control Block table full"
	case 0x56:
		return "Bad buffer address"
	case 0x57:
		return "Duplicate volume"
	case 0x5A:
		return "File structure damaged"
	default:
		return "Unknown error"
	}
}
