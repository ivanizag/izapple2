package apple2sdl

import (
	"github.com/veandco/go-sdl2/sdl"

	"go6502/apple2"
)

// SDLRun starts the Apple2 emulator on SDL
func SDLRun(a *apple2.Apple2) {
	window, renderer, err := sdl.CreateWindowAndRenderer(800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic("Failed to create window")
	}
	defer window.Destroy()
	defer renderer.Destroy()
	window.SetTitle("Apple2")
	renderer.Clear()
	renderer.Present()

	kp := newSDLKeyBoard()
	a.SetKeyboardProvider(&kp)
	go a.Run(false, false)

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				//fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%c\tmodifiers:%d\tstate:%d\trepeat:%d\n",
				//	t.Timestamp, t.Type, t.Keysym.Sym, t.Keysym.Mod, t.State, t.Repeat)
				kp.putKey(t)
			case *sdl.TextInputEvent:
				//fmt.Printf("[%d ms] TextInput\ttype:%d\texts:%s\n",
				//	t.Timestamp, t.Type, t.GetText())
				kp.putText(t)
			}
		}
		sdl.Delay(1000 / 60)
	}

}
