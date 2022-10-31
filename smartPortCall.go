package izapple2

import "fmt"

/*
	A smartPort device
*/

type smartPortDevice interface {
	exec(call *smartPortCall) uint8
}

const (
	proDosDeviceCommandStatus     = 0
	proDosDeviceCommandReadBlock  = 1
	proDosDeviceCommandWriteBlock = 2
	proDosDeviceCommandFormat     = 3
	proDosDeviceCommandControl    = 4
	proDosDeviceCommandInit       = 5
	proDosDeviceCommandOpen       = 6
	proDosDeviceCommandClose      = 7
	proDosDeviceCommandRead       = 8
	proDosDeviceCommandWrite      = 9
)

const (
	prodosDeviceStatusCodeDevice             = 0
	prodosDeviceStatusCodeDeviceControlBlock = 1
	prodosDeviceStatusCodeNewline            = 2
	prodosDeviceStatusCodeDeviceInfo         = 3
)

const (
	prodosDeviceStatusCodeTypeBlock       = uint8(1) << 7
	prodosDeviceStatusCodeTypeWrite       = uint8(1) << 6
	prodosDeviceStatusCodeTypeRead        = uint8(1) << 5
	prodosDeviceStatusCodeTypeOnline      = uint8(1) << 4
	prodosDeviceStatusCodeTypeFormat      = uint8(1) << 3
	prodosDeviceStatusCodeTypeProtected   = uint8(1) << 2
	prodosDeviceStatusCodeTypeInterruping = uint8(1) << 1
	prodosDeviceStatusCodeTypeOpen        = uint8(1) << 0
)

const (
	proDosDeviceNoError             = uint8(0)
	proDosDeviceBadCommand          = uint8(1)
	proDosDeviceErrorIO             = uint8(0x27)
	proDosDeviceErrorNoDevice       = uint8(0x28)
	proDosDeviceErrorWriteProtected = uint8(0x2b)
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
	if spc.command != proDosDeviceCommandStatus {
		panic("Status code paremeter requeted for a non status smartport call")
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
	case proDosDeviceCommandStatus:
		return fmt.Sprintf("STATUS(%v, unit=%v, code=%v)",
			spc.command, spc.unit(),
			spc.statusCode())
	case proDosDeviceCommandReadBlock:
		return fmt.Sprintf("READBLOCK(%v, unit=%v, block=%v)",
			spc.command, spc.unit(),
			spc.param24(4))
	case proDosDeviceCommandWriteBlock:
		return fmt.Sprintf("WRITEBLOCK(%v, unit=%v, block=%v)",
			spc.command, spc.unit(),
			spc.param24(4))
	case proDosDeviceCommandControl:
		return fmt.Sprintf("CONTROL(%v, unit=%v, code=%v)",
			spc.command, spc.unit(),
			spc.param8(4))
	case proDosDeviceCommandInit:
		return fmt.Sprintf("INIT(%v, unit=%v)",
			spc.command, spc.unit())
	case proDosDeviceCommandOpen:
		return fmt.Sprintf("OPEN(%v, unit=%v)",
			spc.command, spc.unit())
	case proDosDeviceCommandClose:
		return fmt.Sprintf("CLOSE(%v, unit=%v)",
			spc.command, spc.unit())
	case proDosDeviceCommandRead:
		return fmt.Sprintf("READ(%v, unit=%v, pos=%v, len=%v)",
			spc.command, spc.unit(),
			spc.param24(6),
			spc.param16(4))
	case proDosDeviceCommandWrite:
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
	case proDosDeviceNoError:
		return "SUCCESS"
	case proDosDeviceBadCommand:
		return "BAD_COMMAND"
	case proDosDeviceErrorIO:
		return "ERROR_IO"
	case proDosDeviceErrorNoDevice:
		return "NO_DEVICE"
	case proDosDeviceErrorWriteProtected:
		return "WRITE_PROTECT_ERROR"
	default:
		return string(code)

	}
}
