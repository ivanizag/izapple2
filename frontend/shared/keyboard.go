package shared

import (
	"fmt"
	"time"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"
)

/*
The keyboard of the frontends with a window: what the user types goes to the
machine, and the keys that command the emulator act on it and on the view.

The frontends only translate the keys of their toolkit, they all answer to the
same ones. The help screen lists them, see HelpMessage.
*/

// Pasting is too fast for the Apple II, so it is slowed down to this per
// character
const pasteDelay = 20 * time.Millisecond

// Keyboard sends what the user types to the machine and runs the commands the
// keys have on the emulator
type Keyboard struct {
	a       *izapple2.Apple2
	channel *izapple2.KeyboardChannel
	view    *View
}

func NewKeyboard(a *izapple2.Apple2, view *View) *Keyboard {
	var k Keyboard
	k.a = a
	k.channel = izapple2.NewKeyboardChannel(a)
	k.view = view
	return &k
}

// PutText sends a text to the machine as the characters it is made of
func (k *Keyboard) PutText(text string) {
	k.channel.PutText(text)
}

// PutChar sends a character to the machine
func (k *Keyboard) PutChar(char uint8) {
	k.channel.PutChar(char)
}

// PutKey acts on a key pressed: it sends the character the Apple II keyboard
// would produce for it, or runs the command it has on the emulator
func (k *Keyboard) PutKey(key Key, ctrl bool, shift bool) {
	if char, ok := CharForKey(key, ctrl); ok {
		k.channel.PutChar(char)
		return
	}

	k.command(key, ctrl, shift)
}

// PutCtrlLetter sends a letter pressed with the control key, and returns
// whether it was a letter at all
func (k *Keyboard) PutCtrlLetter(letter rune) bool {
	char, ok := CharForCtrlLetter(letter)
	if ok {
		k.channel.PutChar(char)
	}
	return ok
}

// Paste types a text on the machine. It returns at once, the text is typed on
// its own goroutine slowly enough for the Apple II to keep up.
func (k *Keyboard) Paste(text string) {
	go func() {
		for _, ch := range text {
			if ch == '\r' || ch == '\n' {
				// Translate the CR/LFs
				k.channel.PutChar(13)
			} else {
				k.channel.PutRune(ch)
			}
			time.Sleep(pasteDelay)
		}
	}()
}

// command runs what a key does to the emulator and to what is shown
func (k *Keyboard) command(key Key, ctrl bool, shift bool) {
	switch key {
	case KeyF1:
		k.view.ToggleHelp()
	case KeyF2:
		// While the help is shown, F2 alone triggers a reset. This is handy
		// when the window manager (for example KDE) captures Ctrl-F2.
		if ctrl || k.view.ShowHelp {
			k.a.SendCommand(izapple2.CommandReset)
			k.view.ShowHelp = false
		}
	case KeyF4:
		k.a.SendCommand(izapple2.CommandToggleCPUTrace)
	case KeyF5:
		if ctrl {
			k.a.SendCommand(izapple2.CommandShowSpeed)
		} else {
			k.a.SendCommand(izapple2.CommandToggleSpeed)
		}
	case KeyF6:
		k.view.NextScreenMode()
	case KeyF7:
		k.view.ShowPages = !k.view.ShowPages
	case KeyF8:
		k.view.ToggleDropTargets()
	case KeyF9:
		k.a.SendCommand(izapple2.CommandDumpDebugInfo)
	case KeyF10:
		if ctrl {
			k.view.ShowCharGen = !k.view.ShowCharGen
		} else if shift {
			k.view.ShowAltText = !k.view.ShowAltText
		} else {
			k.a.SendCommand(izapple2.CommandNextCharGenPage)
		}
	case KeyF12, KeyPrintScreen:
		if ctrl {
			screen.AddScenario(k.a.GetVideoSource(), "../../screen/test_resources/")
		} else {
			k.saveSnapshot()
		}
	case KeyPause:
		k.a.SendCommand(izapple2.CommandPauseUnpause)
	}
}

func (k *Keyboard) saveSnapshot() {
	err := screen.SaveSnapshot(k.a.GetVideoSource(),
		screen.ScreenModeColorScanlines, "snapshot.png")
	if err != nil {
		fmt.Printf("Error saving snapshoot: %v.\n.", err)
	} else {
		fmt.Println("Saving snapshot 'snapshot.png'")
	}
}
