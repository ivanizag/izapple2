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
	now   func() time.Time // The time of the machine
	trace bool
}

// NewSmartPortFujinetClock creates a new fujinet device that gives the time
// of now
func NewSmartPortFujinetClock(now func() time.Time) *SmartPortFujinetClock {
	var d SmartPortFujinetClock
	d.now = now
	return &d
}

func (d *SmartPortFujinetClock) exec(call *smartPortCall) (uint8, []uint8) {
	var result uint8
	var data []uint8

	switch call.command {

	case smartPortCommandOpen:
		result = smartPortNoError

	case smartPortCommandClose:
		result = smartPortNoError

	case smartPortCommandStatus:
		result, data = d.status(call.statusCode())

	default:
		// Prodos device command not supported
		result = smartPortErrorIO
	}

	if d.trace {
		fmt.Printf("[SmartPortFujinetClock] Command %v, return %s \n",
			call, smartPortErrorMessage(result))
	}

	return result, data
}

func (d *SmartPortFujinetClock) status(code uint8) (uint8, []uint8) {
	var data []uint8

	switch code {
	case smartPortStatusCodeDevice:
		// See iwmNetwork::encode_status_reply_packet()
		data = []uint8{
			0,       // NA for a clock
			0, 0, 0, // Block size is 0
		}

	case smartPortStatusCodeDeviceInfo:
		// See iwmNetwork::encode_status_reply_packet()
		data = []uint8{
			smartPortStatusCodeTypeRead & smartPortStatusCodeTypeOnline,
			0, 0, 0, // Block size is 0
			8, 'F', 'N', '_', 'C', 'L', 'O', 'C', 'K', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
			0x13,       // Type fujinet clock.
			0x00,       // Subtype
			0x00, 0x01, // Firmware version
		}

	case 'T':
		// Get time and send it in easy to use format
		now := d.now()
		data = []uint8{
			uint8(now.Year() / 100),
			uint8(now.Year() % 100),
			uint8(now.Month()),
			uint8(now.Day()),
			uint8(now.Hour()),
			uint8(now.Minute()),
			uint8(now.Second()),
		}

	case 'P':
		// Get time and send it in ProDOS format
		// See 6.1 in https://prodos8.com/docs/techref/adding-routines-to-prodos/
		now := d.now()

		datelo := uint8(now.Day()) + uint8(now.Month())<<5
		datehi := uint8(now.Year()%100)<<1 + uint8(now.Month())>>3
		timelo := uint8(now.Minute())
		timehi := uint8(now.Hour())

		data = []uint8{
			datelo,
			datehi,
			timelo,
			timehi,
		}
	}

	return smartPortNoError, data // The return code is always success
}
