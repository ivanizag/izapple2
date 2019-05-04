package apple2

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

/*
Uses the console standard input and output to interface with the machine.
Input is buffered until the next CR. This avoids working in place, a line
for input is added at the end.
Outut is done in place using ANSI escape sequences.

Those tricks do not work with the Apple2e ROM
*/

type ansiConsoleFrontend struct {
	apple2         *Apple2
	keyChannel     chan uint8
	extraLineFeeds chan int
	textUpdated    bool
	stdinKeyboard  bool
}

func newAnsiConsoleFrontend(a *Apple2, stdinKeyboard bool) *ansiConsoleFrontend {
	var fe ansiConsoleFrontend
	fe.apple2 = a
	fe.stdinKeyboard = stdinKeyboard
	fe.subscribeToTextPages()
	return &fe
}

func (fe *ansiConsoleFrontend) subscribeToTextPages() {
	observer := func(_ uint8, _ bool) {
		fe.textUpdated = true
	}
	for i := 0x04; i < 0x08; i++ {
		fe.apple2.mmu.physicalMainRAM[i].observer = observer
	}
}

const refreshDelayMs = 100

func (fe *ansiConsoleFrontend) GetKey(strobed bool) (key uint8, ok bool) {

	// Init the first time
	if fe.keyChannel == nil {
		stdinReader := func(c chan uint8) {
			reader := bufio.NewReader(os.Stdin)
			for {
				key, err := reader.ReadByte()
				if err != nil {
					fmt.Println(err)
					return
				}
				c <- key
			}
		}

		fe.keyChannel = make(chan uint8, 100)
		go stdinReader(fe.keyChannel)
	}

	if !strobed {
		// We must use the strobe to control the flow from stdin
		ok = false
		return
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

func (fe *ansiConsoleFrontend) textModeGoRoutineFast() {
	fe.extraLineFeeds = make(chan int, 100)

	fmt.Printf(strings.Repeat("\n", 26))
	for {
		if fe.textUpdated {
			fe.textUpdated = false
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
			isAltText := fe.apple2.isApple2e && fe.apple2.io.isSoftSwitchActive(ioFlagAltChar)
			var i, j, h, c uint8
			// Top, middle and botton screen
			for i = 0; i < 120; i = i + 40 {
				// Memory pages
				for j = 0x04; j < 0x08; j++ {
					p := fe.apple2.mmu.physicalMainRAM[j]
					// The two half pages
					for _, h = range []uint8{0, 128} {
						line := ""
						for c = i + h; c < i+h+40; c++ {
							line += textMemoryByteToString(p.internalPeek(c), isAltText)
						}
						fmt.Printf("# %v #\n", line)
					}
				}
			}

			fmt.Println(strings.Repeat("#", 44))
			if fe.stdinKeyboard {
				fmt.Print("\033[KLine: ")
			}
		}
		time.Sleep(refreshDelayMs * time.Millisecond)
	}
}

func (fe *ansiConsoleFrontend) textModeGoRoutine() {
	fe.extraLineFeeds = make(chan int, 100)

	fmt.Printf(strings.Repeat("\n", 26))
	for {
		if fe.textUpdated {
			fe.textUpdated = false
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

			pageIndex := 0
			if fe.apple2.io.isSoftSwitchActive(ioFlagSecondPage) {
				pageIndex = 1
			}
			isAltText := fe.apple2.isApple2e && fe.apple2.io.isSoftSwitchActive(ioFlagAltChar)

			fmt.Println(strings.Repeat("#", 44))
			for line := 0; line < 24; line++ {
				text := ""
				for col := 0; col < 40; col++ {
					value := getTextChar(fe.apple2, col, line, pageIndex)
					text += textMemoryByteToString(value, isAltText)
				}
				fmt.Printf("# %v #\n", text)
			}

			fmt.Println(strings.Repeat("#", 44))
			if fe.stdinKeyboard {
				fmt.Print("\033[KLine: ")
			}

			//saveSnapshot(fe.apple2)
		}
		time.Sleep(refreshDelayMs * time.Millisecond)
	}
}

func textMemoryByteToString(value uint8, isAltCharSet bool) string {
	// See https://en.wikipedia.org/wiki/Apple_II_character_set
	// Supports the new lowercase characters in the Apple2e
	// Only ascii from 0x20 to 0x5F is visible
	topBits := value >> 6
	isInverse := topBits == 0
	isFlash := topBits == 1
	if isFlash && isAltCharSet {
		// On the Apple2e with lowercase chars there is not flash mode.
		isFlash = false
		isInverse = true
	}

	if isAltCharSet {
		value = value & 0x7F
	} else {
		value = value & 0x3F
	}

	if value < 0x20 {
		value += 0x40
	}

	if value == 0x7f {
		// DEL is full box
		value = '_'
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

func textMemoryByteToStringHex(value uint8, _ bool) string {
	return fmt.Sprintf("%02x ", value)
}
