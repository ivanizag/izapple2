package apple2

import (
	"bufio"
	"os"
)

type ansiConsoleFrontend struct {
	keyChannel chan uint8
}

func _stdinReader(c chan uint8) {
	reader := bufio.NewReader(os.Stdin)
	for {
		byte, err := reader.ReadByte()
		if err != nil {
			panic(err)
		}
		c <- byte
	}
}

func (fe *ansiConsoleFrontend) getKey() (key uint8, ok bool) {
	if fe.keyChannel == nil {
		fe.keyChannel = make(chan uint8, 100)
		go _stdinReader(fe.keyChannel)
	}

	select {
	case key = <-fe.keyChannel:
		if key == 10 {
			key = 13
		}
		ok = true
	default:
		ok = false
	}
	return
}
