package izapple2

import (
	"fmt"
	"time"
)

/*

A clock device that could be implemented by Fujinet:

*/

// SmartPortFujinetClock represents a Fujinet clock device
type SmartPortFujinetClock struct {
	host  *CardSmartPort // For DMA
	trace bool
}

// NewSmartPortFujinetClock creates a new fujinet device
func NewSmartPortFujinetClock(host *CardSmartPort) *SmartPortFujinetClock {
	var d SmartPortFujinetClock
	d.host = host
	return &d
}

func (d *SmartPortFujinetClock) exec(call *smartPortCall) uint8 {
	var result uint8

	switch call.command {

	case smartPortCommandOpen:
		result = smartPortNoError

	case smartPortCommandClose:
		result = smartPortNoError

	case smartPortCommandStatus:
		address := call.param16(2)
		result = d.status(call.statusCode(), address)

	default:
		// Prodos device command not supported
		result = smartPortErrorIO
	}

	if d.trace {
		fmt.Printf("[SmartPortFujinetClock] Command %v, return %s \n",
			call, smartPortErrorMessage(result))
	}

	return result
}

func (d *SmartPortFujinetClock) status(code uint8, dest uint16) uint8 {

	switch code {
	case smartPortStatusCodeDevice:
		// See iwmNetwork::encode_status_reply_packet()
		d.host.a.mmu.pokeRange(dest, []uint8{
			0,       // NA for a clock
			0, 0, 0, // Block size is 0
		})

	case smartPortStatusCodeDeviceInfo:
		// See iwmNetwork::encode_status_reply_packet()
		d.host.a.mmu.pokeRange(dest, []uint8{
			smartPortStatusCodeTypeRead & smartPortStatusCodeTypeOnline,
			0, 0, 0, // Block size is 0
			8, 'F', 'N', '_', 'C', 'L', 'O', 'C', 'K', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
			0x13,       // Type fujinet clock.
			0x00,       // Subtype
			0x00, 0x01, // Firmware version
		})

	case 'T':
		// Get time and send it in easy to use format
		now := time.Now()
		d.host.a.mmu.pokeRange(dest, []uint8{
			uint8(now.Year() / 100),
			uint8(now.Year() % 100),
			uint8(now.Month()),
			uint8(now.Day()),
			uint8(now.Hour()),
			uint8(now.Minute()),
			uint8(now.Second()),
		})

	case 'P':
		// Get time and send it in ProDOS format
		// See 6.1 in https://prodos8.com/docs/techref/adding-routines-to-prodos/
		now := time.Now()

		datelo := uint8(now.Day()) + uint8(now.Month())<<5
		datehi := uint8(now.Year()%100)<<1 + uint8(now.Month())>>3
		timelo := uint8(now.Minute())
		timehi := uint8(now.Hour())

		d.host.a.mmu.pokeRange(dest, []uint8{
			datelo,
			datehi,
			timelo,
			timehi,
		})
	}

	return smartPortNoError // The return code is always success
}
