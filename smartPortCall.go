package izapple2

import "fmt"

/*
	A smartPort device
*/

type smartPortDevice interface {
	exec(call *smartPortCall) uint8
}

const (
	smartPortCommandStatus     = 0
	smartPortCommandReadBlock  = 1
	smartPortCommandWriteBlock = 2
	smartPortCommandFormat     = 3
	smartPortCommandControl    = 4
	smartPortCommandInit       = 5
	smartPortCommandOpen       = 6
	smartPortCommandClose      = 7
	smartPortCommandRead       = 8
	smartPortCommandWrite      = 9
)

const (
	smartPortStatusCodeDevice             = 0
	smartPortStatusCodeDeviceControlBlock = 1
	smartPortStatusCodeNewline            = 2
	smartPortStatusCodeDeviceInfo         = 3
)

const (
	smartPortStatusCodeTypeBlock       = uint8(1) << 7
	smartPortStatusCodeTypeWrite       = uint8(1) << 6
	smartPortStatusCodeTypeRead        = uint8(1) << 5
	smartPortStatusCodeTypeOnline      = uint8(1) << 4
	smartPortStatusCodeTypeFormat      = uint8(1) << 3
	smartPortStatusCodeTypeProtected   = uint8(1) << 2
	smartPortStatusCodeTypeInterruping = uint8(1) << 1
	smartPortStatusCodeTypeOpen        = uint8(1) << 0
)

const (
	smartPortNoError             = uint8(0)
	smartPortBadCommand          = uint8(1)
	smartPortErrorIO             = uint8(0x27)
	smartPortErrorNoDevice       = uint8(0x28)
	smartPortErrorWriteProtected = uint8(0x2b)
)

type smartPortCall struct {
	host    *CardSmartPort
	command uint8

	address uint16  // When the params are on the Apple memory
	params  []uint8 // When the params are built externally as on a ProDOS to SP translation
}

func newSmartPortCall(host *CardSmartPort, command uint8, address uint16) *smartPortCall {
	var spc smartPortCall
	spc.host = host
	spc.command = command
	spc.address = address
	spc.params = nil
	return &spc
}

func newSmartPortCallSynthetic(host *CardSmartPort, command uint8, params []uint8) *smartPortCall {
	var spc smartPortCall
	spc.host = host
	spc.command = command
	spc.address = 0xffff
	spc.params = params
	return &spc
}

func (spc *smartPortCall) unit() uint8 {
	return spc.param8(1)
}

func (spc *smartPortCall) statusCode() uint8 {
	if spc.command != smartPortCommandStatus {
		panic("Status code paremeter requeted for a non status smartPort call")
	}
	return spc.param8(4)
}

func (spc *smartPortCall) param8(offset uint8) uint8 {
	if spc.params == nil {
		return spc.host.a.mmu.Peek(spc.address + uint16(offset))
	}

	if int(offset) >= len(spc.params) {
		panic("Synthetised smartpot call out of range")
	}

	return spc.params[offset]
}

func (spc *smartPortCall) param16(offset uint8) uint16 {
	return uint16(spc.param8(offset)) +
		uint16(spc.param8(offset+1))<<8
}

func (spc *smartPortCall) param24(offset uint8) uint32 {
	return uint32(spc.param8(offset)) +
		uint32(spc.param8(offset+1))<<8 +
		uint32(spc.param8(offset+2))<<16
}

func (spc *smartPortCall) paramData(offset uint8) []uint8 {
	address := uint16(spc.param8(offset)) +
		uint16(spc.param8(offset+1))<<8

	size := spc.host.a.mmu.peekWord(address)

	data := make([]uint8, size)
	for i := 0; i < int(size); i++ {
		data[i] = spc.host.a.mmu.Peek(address + 2 + uint16(i))
	}

	return data
}

func (spc *smartPortCall) String() string {
	switch spc.command {
	case smartPortCommandStatus:
		return fmt.Sprintf("STATUS(%v, unit=%v, code=%v)",
			spc.command, spc.unit(),
			spc.statusCode())
	case smartPortCommandReadBlock:
		return fmt.Sprintf("READBLOCK(%v, unit=%v, block=%v)",
			spc.command, spc.unit(),
			spc.param24(4))
	case smartPortCommandWriteBlock:
		return fmt.Sprintf("WRITEBLOCK(%v, unit=%v, block=%v)",
			spc.command, spc.unit(),
			spc.param24(4))
	case smartPortCommandControl:
		return fmt.Sprintf("CONTROL(%v, unit=%v, code=%v)",
			spc.command, spc.unit(),
			spc.param8(4))
	case smartPortCommandInit:
		return fmt.Sprintf("INIT(%v, unit=%v)",
			spc.command, spc.unit())
	case smartPortCommandOpen:
		return fmt.Sprintf("OPEN(%v, unit=%v)",
			spc.command, spc.unit())
	case smartPortCommandClose:
		return fmt.Sprintf("CLOSE(%v, unit=%v)",
			spc.command, spc.unit())
	case smartPortCommandRead:
		return fmt.Sprintf("READ(%v, unit=%v, pos=%v, len=%v)",
			spc.command, spc.unit(),
			spc.param24(6),
			spc.param16(4))
	case smartPortCommandWrite:
		return fmt.Sprintf("WRITE(%v, unit=%v, pos=%v, len=%v)",
			spc.command, spc.unit(),
			spc.param24(6),
			spc.param16(4))

	default:
		return fmt.Sprintf("UNKNOWN(%v, unit=%v)",
			spc.command, spc.unit())
	}
}

func smartPortErrorMessage(code uint8) string {
	switch code {
	case smartPortNoError:
		return "SUCCESS"
	case smartPortBadCommand:
		return "BAD_COMMAND"
	case smartPortErrorIO:
		return "ERROR_IO"
	case smartPortErrorNoDevice:
		return "NO_DEVICE"
	case smartPortErrorWriteProtected:
		return "WRITE_PROTECT_ERROR"
	default:
		return string(code)

	}
}
