package apple2

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type ansiConsoleFrontend struct {
	keyChannel     chan uint8
	extraLineFeeds chan int
}

const refreshDelayMs = 100

func (fe *ansiConsoleFrontend) getKey() (key uint8, ok bool) {
	stdinReader := func(c chan uint8) {
		reader := bufio.NewReader(os.Stdin)
		for {
			byte, err := reader.ReadByte()
			if err != nil {
				panic(err)
			}
			c <- byte
		}
	}

	if fe.keyChannel == nil {
		fe.keyChannel = make(chan uint8, 100)
		go stdinReader(fe.keyChannel)
	}

	select {
	case key = <-fe.keyChannel:
		if key == 10 {
			key = 13
			if fe.extraLineFeeds != nil {
				fe.extraLineFeeds <- 1
			}
		}
		ok = true
	default:
		ok = false
	}
	return
}

func ansiCursorUp(steps int) {
	fmt.Printf("\033[%vA", steps)
}

func (fe *ansiConsoleFrontend) textModeGoRoutine(tp *textPages) {
	fe.extraLineFeeds = make(chan int, 100)

	fmt.Printf(strings.Repeat("\n", 26))
	for {
		if tp.strobe() {
			// Go up
			ansiCursorUp(26)
			done := false
			for !done {
				select {
				case lineFeeds := <-fe.extraLineFeeds:
					ansiCursorUp(lineFeeds)
				default:
					done = true
				}
			}

			fmt.Println(strings.Repeat("#", 44))

			// See "Understand the Apple II", page 5-10
			// http://www.applelogic.org/files/UNDERSTANDINGTHEAII.pdf
			var i, j, h uint8
			// Top, middle and botton screen
			for i = 0; i < 120; i = i + 40 {
				// Memory pages
				for _, p := range tp.pages {
					// The two half pages
					for _, h = range []uint8{0, 128} {
						line := ""
						for j = i + h; j < i+h+40; j++ {
							line += textMemoryByteToString(p.Peek(j))
						}
						fmt.Printf("# %v #\n", line)
					}
				}
			}

			fmt.Println(strings.Repeat("#", 44))
			fmt.Print("\033[KLine: ")

		}
		time.Sleep(refreshDelayMs * time.Millisecond)
	}
}

func textMemoryByteToString(value uint8) string {
	// See https://en.wikipedia.org/wiki/Apple_II_character_set
	// Only ascii from 0x20 to 0x5F is visible
	topBits := value >> 6
	isInverse := topBits == 0
	isFlash := topBits == 1

	value = (value & 0x3F)
	if value < 0x20 {
		value += 0x40
	}

	if isFlash {
		if value == ' ' {
			// Flashing space in Apple is the full box. It can't be done with ANSI codes
			value = '_'
		}
		return fmt.Sprintf("\033[5m%v\033[0m", string(value))
	} else if isInverse {
		return fmt.Sprintf("\033[7m%v\033[0m", string(value))
	} else {
		return string(value)
	}
}

func textMemoryByteToStringHex(value uint8) string {
	return fmt.Sprintf("%02x ", value)
}
