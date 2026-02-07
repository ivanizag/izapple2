# izapple2 WebAssembly Frontend

A web browser frontend for the izapple2 Apple II emulator, compiled to WebAssembly.

## Features

- ✅ Runs Apple II emulation in any modern web browser
- ✅ Full keyboard support
- ✅ Audio support (speaker clicks and beeps)
- ✅ Multiple disk loading methods:
  - File picker (local files)
  - Drag and drop
  - Load from URLs
  - Sample disk library
- ✅ Screenshot download
- ✅ Responsive design (desktop, tablet, mobile)
- ✅ Reset and pause controls
- ✅ No installation required

## Requirements

### For Building

- Go 1.16 or later
- Standard build tools (make, bash)

### For Running (Browser)

- Modern web browser with WebAssembly support:
  - Chrome/Edge 57+
  - Firefox 52+
  - Safari 11+
  - Opera 44+

## Quick Start

### Build

```bash
cd frontend/a2wasm
./build.sh
```

This will:
1. Compile Go code to WebAssembly (`web/izapple2.wasm`)
2. Copy the Go WASM runtime (`web/wasm_exec.js`)
3. Create a compressed version (`web/izapple2.wasm.gz`)

### Run Locally

```bash
./serve.sh
```

Then open http://localhost:8080 in your browser.

The emulator will boot to Applesoft BASIC. Use the file picker or drag-and-drop to load disk images.

## Using the Emulator

### Loading Disk Images

**Method 1: File Picker**
1. Click "Load Disk 1" or "Load Disk 2"
2. Select a disk image file (.dsk, .woz, .nib, .po, .2mg, .zip, .gz)
3. The disk will be loaded and the emulator will boot from it

**Method 2: Drag and Drop**
1. Drag a disk image file from your computer
2. Drop it on the emulator screen
3. Multiple files will load into Drive 1 and Drive 2 respectively

**Method 3: URL Loading**
1. Use URL parameters:
   ```
   http://localhost:8080?disk1=https://example.com/disk.dsk
   ```
2. Or load from JavaScript:
   ```javascript
   window.loadDiskFromURL(1, "https://example.com/disk.dsk");
   ```

**Method 4: Sample Disks**
- Click sample disk buttons in the control panel
- These load pre-configured disk images

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| F1 | Show/Hide help |
| Ctrl+F2 | Reset emulator |
| F4 | Toggle CPU trace |
| F5 | Fast/Normal speed |
| Ctrl+F5 | Show speed/FPS |
| F6 | Next screen mode (NTSC, Green, etc.) |
| F7 | Show/Hide screen pages |
| F10 | Next character set |
| Ctrl+F10 | Show/Hide character generator |
| Shift+F10 | Show/Hide alternate text |
| F12 | Save screenshot |
| Pause | Pause/Resume emulation |
| Left Alt | Open-Apple |
| Right Alt | Closed-Apple |

### Controls

**Reset Button**: Resets the Apple II (same as Ctrl+F2)

**Pause Button**: Pauses/resumes emulation

**Screenshot Button**: Captures current screen as PNG and downloads it

## File Structure

```
frontend/a2wasm/
├── main.go           # Entry point and Ebiten game loop
├── keyboard.go       # Keyboard input handling
├── speaker.go        # Audio output
├── disk_loader.go    # JavaScript bridge for disk loading (WASM-only)
├── build.sh          # Build script
├── serve.sh          # Development server script
├── README.md         # This file
└── web/              # Web assets
    ├── index.html    # Main HTML page
    ├── loader.js     # WASM initialization and UI logic
    ├── styles.css    # Styling
    ├── izapple2.wasm # Compiled WASM binary (after build)
    └── wasm_exec.js  # Go WASM runtime (after build)
```

## Deployment

### Static Hosting

The `web/` directory contains everything needed for deployment:

1. Build the WASM binary:
   ```bash
   ./build.sh
   ```

2. Upload the `web/` directory to any static hosting service:
   - GitHub Pages
   - Netlify
   - Vercel
   - AWS S3 + CloudFront
   - Your own web server

### Server Configuration

**MIME Types**

Ensure your server is configured with the correct MIME type for WASM:

```
.wasm -> application/wasm
```

**Compression**

Enable gzip or brotli compression for better performance:

```nginx
# Nginx example
location ~ \.wasm$ {
    gzip_static on;
    add_header Content-Type application/wasm;
}
```

**CORS**

If loading disk images from external URLs, ensure CORS is properly configured on the source server.

### Example: GitHub Pages

1. Copy `web/` contents to `docs/` in your repository
2. Enable GitHub Pages in repository settings
3. Set source to `docs/` folder
4. Access at: `https://yourusername.github.io/yourrepo/`

## Supported Disk Formats

- `.dsk` - DOS 3.3 order disk image
- `.woz` - WOZ disk image format
- `.nib` - Nibble format
- `.po` - ProDOS order disk image
- `.2mg` - 2IMG format (with header)
- `.zip` - Compressed disk images (will extract first .dsk file)
- `.gz` - Gzipped disk images

## Browser Compatibility

### Desktop

| Browser | Version | Status |
|---------|---------|--------|
| Chrome | 57+ | ✅ Fully supported |
| Firefox | 52+ | ✅ Fully supported |
| Safari | 11+ | ✅ Supported |
| Edge | 79+ | ✅ Fully supported |
| Opera | 44+ | ✅ Supported |

### Mobile

| Browser | Status | Notes |
|---------|--------|-------|
| Chrome Android | ✅ Supported | No virtual keyboard yet |
| Safari iOS | ✅ Supported | No virtual keyboard yet |
| Firefox Android | ⚠️ Limited | May have audio issues |

## Performance

- **WASM Size**: ~8-12 MB uncompressed, ~2-3 MB gzipped
- **Load Time**: 2-5 seconds on broadband
- **CPU Usage**: ~50-70% of one core on modern hardware
- **Emulation Speed**: 90-100% of native speed
- **Audio Latency**: ~100-200ms (acceptable for emulation)

## Troubleshooting

### "Error loading WASM"

- Check that `izapple2.wasm` exists in the `web/` directory
- Run `./build.sh` to rebuild
- Check browser console for detailed errors

### "Disk loading failed"

- Ensure disk image format is supported
- Check file is not corrupted
- For URL loading, verify CORS is enabled on the source

### No audio

- Check browser console for audio context errors
- Some browsers require user interaction before playing audio
- Click the emulator screen to enable audio

### Slow performance

- Close other browser tabs
- Try Chrome or Edge for best WebAssembly performance
- Check CPU usage in browser task manager

### Keys not working

- Click on the emulator canvas to focus it
- Check that browser extensions aren't intercepting keys
- Try using on-screen buttons as fallback

## Development

### Building for Development

```bash
# Build
./build.sh

# Serve locally
./serve.sh

# Or use wasmserve (auto-rebuilds)
go run github.com/hajimehoshi/wasmserve@latest .
```

### JavaScript Bridge API

The Go code exports these functions to JavaScript:

```javascript
// Load disk from bytes
window.loadDisk(drive, uint8Array, filename)
// Returns: null on success, error string on failure

// Load disk from URL
window.loadDiskFromURL(drive, url)
// Returns: null on success, error string on failure

// Reset emulator
window.resetEmulator()

// Toggle pause
window.togglePause()

// Download screenshot
window.downloadScreenshot()
```

### Modifying the UI

- Edit `web/index.html` for structure
- Edit `web/styles.css` for styling
- Edit `web/loader.js` for JavaScript logic
- Rebuild not needed for web file changes (just refresh browser)

## Known Limitations

1. **No virtual keyboard** - Mobile devices need external keyboard
2. **No save states** - Page refresh loses state (planned for future)
3. **No gamepad** - Only keyboard and mouse supported
4. **File system access** - Can't modify loaded disks (read-only)

## Future Enhancements

- Virtual keyboard overlay for mobile
- Touch joystick support
- Save/load state with IndexedDB
- PWA support (offline capability)
- Multiplayer over WebRTC
- GamePad API support

## Credits

- [izapple2](https://github.com/ivanizag/izapple2) by Ivan Izaguirre
- [Ebiten](https://ebiten.org/) game engine
- Go WebAssembly support

## License

Same as the parent izapple2 project.
