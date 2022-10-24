package izapple2

import (
	"fmt"
)

/*

The network device as implemented by Fujinet:

See:
	https://github.com/FujiNetWIFI/fujinet-platformio/blob/master/lib/device/iwm/network.cpp

*/

// SmartPortFujinet represents a Fujinet device
type SmartPortFujinet struct {
	host  *CardSmartPort // For DMA
	trace bool
}

// NewSmartPortFujinet creates a new fujinet device
func NewSmartPortFujinet(host *CardSmartPort) *SmartPortFujinet {
	var d SmartPortFujinet
	d.host = host

	return &d
}

func (d *SmartPortFujinet) exec(call *smartPortCall) uint8 {
	var result uint8

	switch call.command {
	case proDosDeviceCommandStatus:
		address := call.param16(2)
		result = d.status(call.statusCode(), address)

	case proDosDeviceCommandRead:
		address := call.param16(2)
		block := call.param24(4)
		result = d.readBlock(block, address)

	case proDosDeviceCommandWrite:
		address := call.param16(2)
		block := call.param24(4)
		result = d.writeBlock(block, address)

	default:
		// Prodos device command not supported
		result = proDosDeviceErrorIO
	}

	if d.trace {
		fmt.Printf("[SmartPortFujinet] Command %v, return %s \n",
			call, smartPortErrorMessage(result))
	}

	return result
}

func (d *SmartPortFujinet) readBlock(block uint32, dest uint16) uint8 {
	if d.trace {
		fmt.Printf("[SmartPortFujinet] Read block %v into $%x.\n", block, dest)
	}

	// TODO

	return proDosDeviceNoError
}

func (d *SmartPortFujinet) writeBlock(block uint32, source uint16) uint8 {
	if d.trace {
		fmt.Printf("[SmartPortFujinet] Write block %v from $%x.\n", block, source)
	}

	// TODO

	return proDosDeviceNoError
}

func (d *SmartPortFujinet) status(code uint8, dest uint16) uint8 {

	switch code {
	case prodosDeviceStatusCodeDevice:
		// See iwmNetwork::encode_status_reply_packet()
		d.host.a.mmu.Poke(dest, prodosDeviceStatusCodeTypeRead&prodosDeviceStatusCodeTypeOnline)
		d.host.a.mmu.Poke(dest+1, 0x00)
		d.host.a.mmu.Poke(dest+2, 0x00)
		d.host.a.mmu.Poke(dest+3, 0x00) // Block size is 0

	case prodosDeviceStatusCodeDeviceInfo:
		// See iwmNetwork::encode_status_reply_packet()
		d.host.a.mmu.Poke(dest, prodosDeviceStatusCodeTypeRead&prodosDeviceStatusCodeTypeOnline)
		d.host.a.mmu.Poke(dest+1, 0x00)
		d.host.a.mmu.Poke(dest+2, 0x00)
		d.host.a.mmu.Poke(dest+3, 0x00) // Block size is 0
		d.host.a.mmu.Poke(dest+4, 0x07) // Name length
		d.host.a.mmu.Poke(dest+5, 'N')
		d.host.a.mmu.Poke(dest+6, 'E')
		d.host.a.mmu.Poke(dest+7, 'T')
		d.host.a.mmu.Poke(dest+8, 'W')
		d.host.a.mmu.Poke(dest+9, 'O')
		d.host.a.mmu.Poke(dest+10, 'R')
		d.host.a.mmu.Poke(dest+11, 'K')
		d.host.a.mmu.Poke(dest+12, ' ')
		d.host.a.mmu.Poke(dest+13, ' ')
		d.host.a.mmu.Poke(dest+14, ' ')
		d.host.a.mmu.Poke(dest+15, ' ')
		d.host.a.mmu.Poke(dest+16, ' ')
		d.host.a.mmu.Poke(dest+17, ' ')
		d.host.a.mmu.Poke(dest+18, ' ')
		d.host.a.mmu.Poke(dest+19, ' ')
		d.host.a.mmu.Poke(dest+20, ' ')
		d.host.a.mmu.Poke(dest+20, 0x02) // Type hard disk
		d.host.a.mmu.Poke(dest+20, 0x00) // Subtype network (comment in network.cpp has 0x0a)
		d.host.a.mmu.Poke(dest+23, 0x00)
		d.host.a.mmu.Poke(dest+24, 0x01) // Firmware
	}

	return proDosDeviceNoError // The return code is always success
}
