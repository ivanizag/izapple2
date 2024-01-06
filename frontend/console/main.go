package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	apple2 "github.com/ivanizag/izapple2"
)

func main() {
	a, err := apple2.CreateConfiguredApple()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fe := &ansiConsoleFrontend{}
	a.SetKeyboardProvider(fe)
	go fe.textModeGoRoutine(a)

	a.Run()
}

/*
Uses the console standard input and output to interface with the machine.
Input is buffered until the next CR. This avoids working in place, a line
for input is added at the end.
Outut is done in place using ANSI escape sequences.

Those tricks do not work with the Apple2e ROM
*/

type ansiConsoleFrontend struct {
	keyChannel     chan uint8
	extraLineFeeds chan int
	lastContent    string
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

				if key == 10 {
					key = 13
					if fe.extraLineFeeds != nil {
						fe.extraLineFeeds <- 1
					}
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
		ok = true
	default:
		ok = false
	}
	return
}

func ansiCursorUp(steps int) string {
	return fmt.Sprintf("\033[%vA", steps)
}

func (fe *ansiConsoleFrontend) textModeGoRoutine(a *apple2.Apple2) {
	fe.extraLineFeeds = make(chan int, 100)

	fmt.Printf(strings.Repeat("\n", 24+3))
	for {
		// Go up
		content := ansiCursorUp(24 + 3)
		done := false
		for !done {
			select {
			case lineFeeds := <-fe.extraLineFeeds:
				content += ansiCursorUp(lineFeeds)
			default:
				done = true
			}
		}

		content += apple2.DumpTextModeAnsi(a)
		content += "\033[KLine: "

		if content != fe.lastContent {
			fmt.Print(content)
			fe.lastContent = content
		}
		time.Sleep(refreshDelayMs * time.Millisecond)
	}
}
