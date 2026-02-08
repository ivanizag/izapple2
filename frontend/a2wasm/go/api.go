//go:build js

package main

import (
	"bytes"
	"fmt"
	"image/png"
	"syscall/js"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"
)

var globalApple *izapple2.Apple2

// setupAPI exports functions to JavaScript for React to call
func setupAPI(a *izapple2.Apple2, game *Game) {
	globalApple = a

	api := make(map[string]interface{})

	// Core control
	api["reset"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.SendCommand(izapple2.CommandReset)
		return nil
	})

	api["pause"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.SendCommand(izapple2.CommandPause)
		return nil
	})

	api["resume"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.SendCommand(izapple2.CommandStart)
		return nil
	})

	// Disk operations
	api["loadDisk"] = js.FuncOf(loadDiskJS)
	api["loadDiskFromURL"] = js.FuncOf(loadDiskFromURLJS)

	// Input
	api["sendKey"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		keyCode := uint8(args[0].Int())
		game.keyChannel.PutChar(keyCode)
		return nil
	})

	api["sendText"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		text := args[0].String()
		game.keyChannel.PutText(text)
		return nil
	})

	// State queries
	api["isPaused"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return a.IsPaused()
	})

	api["getFrequency"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return a.GetCurrentFreqMHz()
	})

	api["getDiskInfo"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		// For now, return null (disk info tracking would need additional implementation)
		return nil
	})

	// Configuration
	api["toggleSpeed"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.SendCommand(izapple2.CommandToggleSpeed)
		return nil
	})

	api["setScreenMode"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		mode := args[0].String()
		switch mode {
		case "ntsc":
			game.screenMode = screen.ScreenModeNTSC
		case "plain":
			game.screenMode = screen.ScreenModePlain
		case "green":
			game.screenMode = screen.ScreenModeGreen
		case "amber":
			// Amber not available, use green as fallback
			game.screenMode = screen.ScreenModeGreen
		}
		return nil
	})

	// Screenshot
	api["screenshot"] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		saveScreenshot(a, game.screenMode)
		return nil
	})

	// Export API to JavaScript
	js.Global().Set("wasmAPI", api)

	fmt.Println("WASM API initialized and exported to window.wasmAPI")
}

// loadDiskJS loads a disk from bytes passed from JavaScript
func loadDiskJS(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return "Error: loadDisk requires 3 arguments: drive, bytes, filename"
	}

	drive := args[0].Int()
	dataJS := args[1]
	filename := args[2].String()

	if drive < 1 || drive > 2 {
		return fmt.Sprintf("Error: drive must be 1 or 2, got %d", drive)
	}

	length := dataJS.Get("length").Int()
	if length == 0 {
		return "Error: empty disk image"
	}

	data := make([]byte, length)
	js.CopyBytesToGo(data, dataJS)

	fmt.Printf("Loading disk %d: %s (%d bytes)\n", drive, filename, length)

	diskette, err := izapple2.LoadDisketteFromBytes(data, filename, false)
	if err != nil {
		return fmt.Sprintf("Error loading disk: %v", err)
	}

	err = globalApple.InsertDiskette(drive-1, diskette, filename)
	if err != nil {
		return fmt.Sprintf("Error inserting disk: %v", err)
	}

	fmt.Printf("Successfully loaded disk %d: %s\n", drive, filename)
	return nil
}

// loadDiskFromURLJS loads a disk from a URL
func loadDiskFromURLJS(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return "Error: loadDiskFromURL requires 2 arguments: drive, url"
	}

	drive := args[0].Int()
	url := args[1].String()

	if drive < 1 || drive > 2 {
		return fmt.Sprintf("Error: drive must be 1 or 2, got %d", drive)
	}

	fmt.Printf("Loading disk %d from URL: %s\n", drive, url)

	diskette, err := izapple2.LoadDiskette(url)
	if err != nil {
		return fmt.Sprintf("Error loading disk from URL: %v", err)
	}

	filename := url[len(url)-20:]
	if len(url) < 20 {
		filename = url
	}

	err = globalApple.InsertDiskette(drive-1, diskette, filename)
	if err != nil {
		return fmt.Sprintf("Error inserting disk: %v", err)
	}

	fmt.Printf("Successfully loaded disk %d from URL: %s\n", drive, url)
	return nil
}

// saveScreenshot captures a screenshot and triggers download via JavaScript
func saveScreenshot(a *izapple2.Apple2, screenMode int) {
	if a == nil {
		return
	}

	vs := a.GetVideoSource()
	img := screen.Snapshot(vs, screenMode)
	if img == nil {
		fmt.Println("Error: failed to capture screenshot")
		return
	}

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		fmt.Printf("Error encoding screenshot: %v\n", err)
		return
	}
	pngData := buf.Bytes()

	dst := js.Global().Get("Uint8Array").New(len(pngData))
	js.CopyBytesToJS(dst, pngData)

	js.Global().Call("triggerDownload", dst, "screenshot.png")

	fmt.Println("Screenshot saved")
}
