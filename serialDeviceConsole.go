package izapple2

import (
	"bufio"
	"os"
	"sync"
)

/*
serialDeviceConsole writes to the standard output and reads the standard
input. The output is text: the high bit is removed and the carriage returns
are new lines. The new lines of the input are carriage returns.
*/
type serialDeviceConsole struct {
	received chan uint8
	once     sync.Once
}

func newSerialDeviceConsole() *serialDeviceConsole {
	return &serialDeviceConsole{
		received: make(chan uint8, serialReceiveBuffer),
	}
}

func (d *serialDeviceConsole) Send(value uint8) {
	value &= 0x7f
	if value == '\r' {
		value = '\n'
	}
	os.Stdout.Write([]byte{value})
}

func (d *serialDeviceConsole) Receive() (uint8, bool) {
	// The standard input is read once something is asked
	d.once.Do(func() {
		go func() {
			reader := bufio.NewReader(os.Stdin)
			for {
				value, err := reader.ReadByte()
				if err != nil {
					return
				}
				if value == '\n' {
					value = '\r'
				}
				d.received <- value
			}
		}()
	})
	select {
	case value := <-d.received:
		return value, true
	default:
		return 0, false
	}
}

func (d *serialDeviceConsole) isReady() bool {
	return true
}
