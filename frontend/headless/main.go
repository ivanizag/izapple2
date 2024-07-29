package main

import (
	"bufio"
	"fmt"
	"image/gif"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"
)

func main() {
	a, err := izapple2.CreateConfiguredApple()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fe := &headLessFrontend{}
	fe.keyChannel = make(chan uint8, 200)
	a.SetKeyboardProvider(fe)
	go a.Start(true /*paused*/)

	inReader := bufio.NewReader(os.Stdin)
	done := false
	for !done {
		fmt.Print("* ")
		text, err := inReader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		text = strings.TrimSpace(text)
		parts := strings.Split(text, " ")
		command := strings.ToLower(parts[0])
		switch command {

		// General commands
		case "quit":
			a.SendCommand(izapple2.CommandKill)
			done = true
		case "help":
			fmt.Print(help)

		// Emulation control commands
		case "start":
			a.SendCommand(izapple2.CommandStart)
			spinWait(func() bool { return !a.IsPaused() })
		case "pause":
			a.SendCommand(izapple2.CommandPause)
			spinWait(func() bool { return a.IsPaused() })
		case "run":
			if len(parts) != 2 {
				fmt.Printf("Usage: run <cycles>\n")
			} else if cycles, err := strconv.Atoi(parts[1]); err != nil {
				fmt.Printf("Usage: run <cycles>\n")
			} else if !a.IsPaused() {
				fmt.Printf("Emulation is already running\n")
			} else {
				a.RequestFastMode()
				a.SetCycleBreakpoint(a.GetCycles() + uint64(cycles)*1000)
				a.SendCommand(izapple2.CommandStart)
				spinWait(func() bool { return a.BreakPoint() })
				a.ReleaseFastMode()
			}
		case "cycle":
			fmt.Printf("%v\n", a.GetCycles())
		case "reset":
			a.SendCommand(izapple2.CommandReset)

		// Keyboard related commands
		case "key":
			if len(parts) < 2 {
				fmt.Println("Usage: key <number>")
			} else {
				code, err := strconv.Atoi(parts[1])
				if err != nil || code < 0 || code > 127 {
					fmt.Println("Usage: key <number from 0 to 127>")
				} else {
					fe.putKey(uint8(code))
				}
			}
		case "type":
			text := strings.Join(parts[1:], " ")
			for _, char := range text {
				fe.putKey(uint8(char))
			}
		case "enter":
			fe.putKey(13)
		case "clearkeys":
			fe.clearKeyQueue()

		// Screen related commands
		case "text":
			fmt.Print(izapple2.DumpTextModeAnsi(a))

		// Old:
		case "png":
			err := screen.SaveSnapshot(a.GetVideoSource(), screen.ScreenModeNTSC, "snapshot.png")
			if err != nil {
				fmt.Printf("Error saving screen: %v.\n.", err)
			} else {
				fmt.Println("Saving screen 'snapshot.png'")
			}

		case "pngm":
			err := screen.SaveSnapshot(a.GetVideoSource(), screen.ScreenModePlain, "snapshot.png")
			if err != nil {
				fmt.Printf("Error saving screen: %v.\n.", err)
			} else {
				fmt.Println("Saving screen 'snapshot.png'")
			}

		case "gif":
			SaveGif(a, "snapshot.gif")

		default:
			fmt.Println("Unknown command.")
		}
	}
}

var help = `
General commands:
	quit
		Quits
	help
		Prints this help

Emulation control commands:
	start
		Runs the emulator
	stop
		Stops the emulator
	run <cycles>
		Runs the emulator for <cycles> thousand cycles at full speed. Waits until completed.
	cycle
		Prints the current cycle count
	reset
		Sends a reset to the emulator

Keyboard related commands:
	key <key>
		Queues the key to the emulator. <key> is a decimal number from 0 to 127.
	type <string>
		Queues the string characters to the emulator. No quotes for the argument, it can have spaces.
	enter
		Queues the enter key to the emulator. Alias for "key 13".
	clearkeys
		Clears the key queue.

Screen related commands:
	text
		Prints the text mode screen.
	* png <filename>
		Stores the active screen to <filename> in PNG format as NTSC color.
	* pngm <filename>
		Same as "png" in monochrome.
	* gif <filename> <seconds> <delay>
		Stores the running screen to <filename> in GIF format during <seconds> with a <delay> per frame
		in 100ths of a second as NTSC color.
		If the emulators is stopped. It is run at full speed during <seconds> and the stopped again.
	* gifm <filename> <seconds> <delay>
		Same as "gif" in monochrome.

`

/*
TODO:
	floppy related commands: load disk....
	joystick related commands: set paddle and button state, dump state
*/

func spinWait(f func() bool) {
	for !f() {
		time.Sleep(time.Millisecond * 1)
	}
}

func SaveGif(a *izapple2.Apple2, filename string) error {
	animation := gif.GIF{}

	delay := 50 * time.Millisecond
	delayHundredsS := 5
	frames := 20 // 1 second

	planned := time.Now()
	for i := 0; i < frames; i++ {
		lapse := time.Until(planned)
		fmt.Printf("%v\n", lapse)
		if lapse > 0 {
			time.Sleep(lapse)
		}

		fmt.Printf("%v\n", time.Now())
		img := screen.SnapshotPaletted(a.GetVideoSource(), screen.ScreenModeNTSC)
		animation.Image = append(animation.Image, img)
		animation.Delay = append(animation.Delay, delayHundredsS)

		planned = planned.Add(delay)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	gif.EncodeAll(f, &animation)
	return nil

}

/*
Uses the console to send commands and queries to an emulated machine.
*/
type headLessFrontend struct {
	keyChannel chan uint8
}

func (fe *headLessFrontend) GetKey(strobed bool) (key uint8, ok bool) {
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

func (fe *headLessFrontend) putKey(key uint8) {
	fe.keyChannel <- key
}

func (fe *headLessFrontend) clearKeyQueue() {
	empty := false
	for !empty {
		select {
		case <-fe.keyChannel:
		default:
			empty = true
		}
	}
}
