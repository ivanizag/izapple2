package izapple2

import (
	"fmt"
	"net"
	"sync"
)

/*
serialDeviceTCP listens on a TCP port, with telnet or nc to connect. There
is one connection at a time, the next one waits for it to be closed. The
port is ready while there is a connection. What is sent without it is lost.
*/
type serialDeviceTCP struct {
	received chan uint8
	mutex    sync.Mutex
	conn     net.Conn
}

func newSerialDeviceTCP(port string) (*serialDeviceTCP, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("the serial port can't listen on TCP port %v: %w", port, err)
	}
	d := &serialDeviceTCP{
		received: make(chan uint8, serialReceiveBuffer),
	}
	go d.accept(listener)
	return d, nil
}

func (d *serialDeviceTCP) accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		d.mutex.Lock()
		d.conn = conn
		d.mutex.Unlock()

		// Read until the connection is closed, then accept the next
		buffer := make([]byte, 256)
		for {
			n, err := conn.Read(buffer)
			for _, value := range buffer[:n] {
				d.received <- value
			}
			if err != nil {
				break
			}
		}

		d.mutex.Lock()
		d.conn = nil
		d.mutex.Unlock()
		conn.Close()
	}
}

func (d *serialDeviceTCP) Send(value uint8) {
	d.mutex.Lock()
	conn := d.conn
	d.mutex.Unlock()
	if conn != nil {
		conn.Write([]byte{value})
	}
}

func (d *serialDeviceTCP) Receive() (uint8, bool) {
	select {
	case value := <-d.received:
		return value, true
	default:
		return 0, false
	}
}

func (d *serialDeviceTCP) isReady() bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.conn != nil
}
