package izapple2

// serialDeviceNone is an empty port
type serialDeviceNone struct{}

func (d *serialDeviceNone) Send(uint8) {}

func (d *serialDeviceNone) Receive() (uint8, bool) {
	return 0, false
}

func (d *serialDeviceNone) isReady() bool {
	return false
}
