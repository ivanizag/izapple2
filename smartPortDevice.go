package izapple2

/*
	A smartport device
*/

type smartPortDevice interface {
	exec(command uint8, params []uint8) uint8
}

const (
	proDosDeviceCommandStatus = 0
	proDosDeviceCommandRead   = 1
	proDosDeviceCommandWrite  = 2
	proDosDeviceCommandFormat = 3
)

const (
	proDosDeviceNoError             = uint8(0)
	proDosDeviceErrorIO             = uint8(0x27)
	proDosDeviceErrorNoDevice       = uint8(0x28)
	proDosDeviceErrorWriteProtected = uint8(0x2b)
)

/*
func smartPortParam8(params []uint8, offset uint8) uint8 {
	if int(offset) >= len(params) {
		return 0
	}
	return params[offset]
}
*/

func smartPortParam16(params []uint8, offset uint8) uint16 {
	if int(offset+1) >= len(params) {
		return 0
	}
	return uint16(params[offset]) + uint16(params[offset+1])<<8
}

func smartPortParam24(params []uint8, offset uint8) uint32 {
	if int(offset+2) >= len(params) {
		return 0
	}
	return uint32(params[offset]) + uint32(params[offset+1])<<8 + uint32(params[offset+2])<<16
}
