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

// setupDiskLoader exports functions to JavaScript for disk loading
func setupDiskLoader(a *izapple2.Apple2) {
	globalApple = a

	// Export loadDisk function to JavaScript
	js.Global().Set("loadDisk", js.FuncOf(loadDiskJS))

	// Export loadDiskFromURL function to JavaScript
	js.Global().Set("loadDiskFromURL", js.FuncOf(loadDiskFromURLJS))

	// Export reset function to JavaScript
	js.Global().Set("resetEmulator", js.FuncOf(resetEmulatorJS))

	// Export pause/unpause function to JavaScript
	js.Global().Set("togglePause", js.FuncOf(togglePauseJS))

	// Export screenshot download function
	js.Global().Set("downloadScreenshot", js.FuncOf(downloadScreenshotJS))

	fmt.Println("Disk loader JavaScript bridge initialized")
}

// loadDiskJS loads a disk from bytes passed from JavaScript
// Args: drive (int), bytes (Uint8Array), filename (string)
// Returns: null on success, error string on failure
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

	// Get the length of the JavaScript Uint8Array
	length := dataJS.Get("length").Int()
	if length == 0 {
		return "Error: empty disk image"
	}

	// Copy JavaScript bytes to Go
	data := make([]byte, length)
	js.CopyBytesToGo(data, dataJS)

	fmt.Printf("Loading disk %d: %s (%d bytes)\n", drive, filename, length)

	// Load diskette using the existing infrastructure
	diskette, err := izapple2.LoadDisketteFromBytes(data, filename, false)
	if err != nil {
		return fmt.Sprintf("Error loading disk: %v", err)
	}

	// Insert disk into the appropriate drive (unit 0 = drive 1, unit 1 = drive 2)
	err = globalApple.InsertDiskette(drive-1, diskette, filename)
	if err != nil {
		return fmt.Sprintf("Error inserting disk: %v", err)
	}

	fmt.Printf("Successfully loaded disk %d: %s\n", drive, filename)
	return nil
}

// loadDiskFromURLJS loads a disk from a URL
// Args: drive (int), url (string)
// Returns: null on success, error string on failure
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

	// Load diskette from URL using existing HTTP support
	diskette, err := izapple2.LoadDiskette(url)
	if err != nil {
		return fmt.Sprintf("Error loading disk from URL: %v", err)
	}

	// Insert disk into the appropriate drive (unit 0 = drive 1, unit 1 = drive 2)
	filename := url[len(url)-20:] // Use last part of URL as name
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

// resetEmulatorJS resets the emulator
func resetEmulatorJS(this js.Value, args []js.Value) interface{} {
	if globalApple != nil {
		globalApple.SendCommand(izapple2.CommandReset)
		fmt.Println("Emulator reset")
	}
	return nil
}

// togglePauseJS toggles pause/unpause
func togglePauseJS(this js.Value, args []js.Value) interface{} {
	if globalApple != nil {
		globalApple.SendCommand(izapple2.CommandPauseUnpause)
		if globalApple.IsPaused() {
			fmt.Println("Emulator paused")
		} else {
			fmt.Println("Emulator unpaused")
		}
	}
	return nil
}

// downloadScreenshotJS captures a screenshot and triggers download via JavaScript
func downloadScreenshotJS(this js.Value, args []js.Value) interface{} {
	if globalApple == nil {
		return "Error: emulator not initialized"
	}

	vs := globalApple.GetVideoSource()
	img := screen.Snapshot(vs, screen.ScreenModeNTSC)
	if img == nil {
		return "Error: failed to capture screenshot"
	}

	// Convert image to PNG bytes
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return fmt.Sprintf("Error encoding screenshot: %v", err)
	}
	pngData := buf.Bytes()

	// Copy to JavaScript
	dst := js.Global().Get("Uint8Array").New(len(pngData))
	js.CopyBytesToJS(dst, pngData)

	// Call JavaScript download function
	js.Global().Call("triggerDownload", dst, "screenshot.png")

	fmt.Println("Screenshot saved")
	return nil
}

// saveScreenshot is called from keyboard.go for F12 key
func saveScreenshot(a *izapple2.Apple2) {
	if a == nil {
		return
	}

	vs := a.GetVideoSource()
	img := screen.Snapshot(vs, screen.ScreenModeNTSC)
	if img == nil {
		fmt.Println("Error: failed to capture screenshot")
		return
	}

	// Convert image to PNG bytes
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		fmt.Printf("Error encoding screenshot: %v\n", err)
		return
	}
	pngData := buf.Bytes()

	// Copy to JavaScript
	dst := js.Global().Get("Uint8Array").New(len(pngData))
	js.CopyBytesToJS(dst, pngData)

	// Call JavaScript download function
	js.Global().Call("triggerDownload", dst, "screenshot.png")

	fmt.Println("Screenshot saved")
}
