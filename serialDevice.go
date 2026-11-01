package izapple2

import (
	"fmt"
	"strconv"
)

/*
The devices of the host on the other side of a serial port:

	none:     nothing connected, the port is never ready
	console:  the standard output and input of the host
	file:     appends what is sent to a file, nothing is received
	tcp:      listens on a TCP port, for one connection at a time
	netmodem: a Hayes modem that dials host:port on the internet

A device is ready, with DCD and DSR low, when it can take what is sent. The
firmware of the cards waits for it before sending.
*/
type serialDevice interface {
	Send(value uint8)
	Receive() (value uint8, ok bool)
	isReady() bool
}

var serialDeviceParams = []paramSpec{
	{"device", "Device on the serial port: none, console, file, tcp or netmodem", "none"},
	{"file", "File to append what is sent with device=file. Defaults to 'serial<slot>.out'", ""},
	{"port", "TCP port to listen on with device=tcp. Defaults to 1976 + slot", ""},
}

const serialReceiveBuffer = 4096

// buildSerialDevice creates the device named on the parameters of a card
// on a slot
func buildSerialDevice(params map[string]string, slot int) (serialDevice, error) {
	device := paramsGetString(params, "device")
	switch device {
	case "none":
		return &serialDeviceNone{}, nil
	case "console":
		return newSerialDeviceConsole(), nil
	case "file":
		filename := paramsGetPath(params, "file")
		if filename == "" {
			filename = fmt.Sprintf("serial%v.out", slot)
		}
		return newSerialDeviceFile(filename)
	case "tcp":
		port := paramsGetString(params, "port")
		if port == "" {
			port = strconv.Itoa(1976 + slot)
		}
		return newSerialDeviceTCP(port)
	case "netmodem":
		return newSerialDeviceNetModem(), nil
	default:
		return nil, fmt.Errorf("unknown serial device '%s', it can be none, console, file, tcp or netmodem", device)
	}
}
