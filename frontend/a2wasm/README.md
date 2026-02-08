# izapple2 WebAssembly Frontend

A modern React + TypeScript frontend for the izapple2 Apple II emulator, compiled to WebAssembly.

## Features

- ✅ Runs Apple II emulation in any modern web browser
- ✅ Modern React + TypeScript + Material-UI interface
- ✅ Full keyboard support (letters, numbers, special keys, arrows)
- ✅ Audio support (speaker clicks and beeps)
- ✅ Multiple disk loading methods:
  - File picker (local files)
  - Drag and drop
  - Load from URLs
- ✅ Emulator controls:
  - Pause/Resume
  - Reset
  - Speed toggle
  - Screen mode selection (NTSC, Plain, Green)
  - Screenshot download
- ✅ Real-time status display (CPU frequency, disk status)
- ✅ Responsive design
- ✅ No installation required

## Technology Stack

- **Frontend**: React 18 + TypeScript + Vite
- **UI Framework**: Material-UI (MUI) 5
- **Emulation Core**: Go + Ebiten (compiled to WebAssembly)
- **Build Tools**: Vite, Go 1.23+

## Requirements

### For Building

- Go 1.23 or later
- Node.js 18+ and npm
- Bash (for build script)

### For Running (Browser)

- Modern web browser with WebAssembly support:
  - Chrome/Edge 79+
  - Firefox 52+
  - Safari 11+
  - Opera 44+

## Quick Start

### Initial Setup

```bash
cd frontend/a2wasm

# Install npm dependencies
npm install

# Build the WASM binary
./build-wasm.sh
```

### Development

```bash
# Start Vite dev server (with hot reload)
npm run dev
```

Then open http://localhost:5173 in your browser.

The emulator will boot to the `]` prompt. You can then load disk images or start typing commands.

### Production Build

```bash
# Build WASM
./build-wasm.sh

# Build React app for production
npm run build

# Preview production build
npm run preview
```

The production files will be in `dist/`.

## Using the Emulator

### Loading Disk Images

**Method 1: File Picker**
1. Click "Choose File" under Drive 1 or Drive 2
2. Select a disk image file (.dsk, .woz, .nib, .po, .2mg, .zip, .gz)
3. The disk will be loaded into the selected drive

**Method 2: Drag and Drop**
1. Drag a disk image file from your computer
2. Drop it on the disk manager panel
3. The first file loads into Drive 1, second into Drive 2

**Method 3: URL Loading**
- Enter a URL in the disk manager (feature available via UI)
- Or use JavaScript:
  ```javascript
  window.wasmAPI.loadDiskFromURL(1, "https://example.com/disk.dsk");
  ```

### Controls

**Top Control Panel**:
- **Pause/Resume**: Freeze/unfreeze emulation
- **Reset**: Reboot the Apple II
- **Toggle Speed**: Switch between normal and fast speed
- **Screenshot**: Download current screen as PNG
- **Screen Mode**: Select display mode (NTSC, Plain, Green)

**Status Bar** (bottom):
- Shows CPU frequency in MHz
- Displays pause state
- Shows keyboard help text

### Keyboard

All standard Apple II keys work:
- Letters, numbers, symbols
- Enter, Escape, Backspace
- Arrow keys (Up, Down, Left, Right)
- Tab, Delete

The emulator uses Ebiten's native keyboard handling for accurate input.

## File Structure

```
frontend/a2wasm/
├── go/                      # Go source code
│   ├── main.go             # Ebiten game loop, keyboard handling
│   ├── api.go              # JavaScript API exports
│   ├── speaker.go          # Audio output
│   └── disk_loader.go      # Disk loading (placeholder)
├── src/                     # React/TypeScript source
│   ├── main.tsx            # React entry point
│   ├── App.tsx             # Main app component
│   ├── api/                # WASM API client
│   │   ├── types.ts        # TypeScript type definitions
│   │   ├── emulator.ts     # Emulator API wrapper
│   │   └── wasm-loader.ts  # WASM initialization
│   ├── components/         # React components
│   │   ├── EmulatorScreen.tsx  # Canvas container
│   │   ├── ControlPanel.tsx    # Top controls
│   │   ├── DiskManager.tsx     # Disk operations UI
│   │   ├── StatusBar.tsx       # Bottom status display
│   │   └── KeyboardHandler.tsx # Keyboard event capture
│   ├── hooks/              # Custom React hooks
│   │   ├── useEmulator.ts  # Emulator state management
│   │   └── useDiskLoader.ts # Disk loading logic
│   └── styles/             # Styling
│       └── theme.ts        # MUI theme customization
├── public/                 # Static assets
│   └── wasm/              # WASM build output
│       ├── izapple2.wasm   # Compiled WASM binary
│       ├── izapple2.wasm.gz # Compressed version
│       └── wasm_exec.js    # Go WASM runtime
├── index.html             # HTML entry point
├── vite.config.ts         # Vite configuration
├── tsconfig.json          # TypeScript configuration
├── package.json           # npm dependencies
├── build-wasm.sh          # WASM build script
└── README.md              # This file
```

## Architecture

```
┌─────────────────────────────────────────┐
│        React + TypeScript UI            │
│  (Controls, Disk Manager, Status Bar)   │
├─────────────────────────────────────────┤
│        TypeScript API Client            │
│         (wasmAPI wrapper)               │
├─────────────────────────────────────────┤
│           Go/WASM Bridge                │
│      (Exported JS functions)            │
├─────────────────────────────────────────┤
│   Go Emulation + Ebiten Rendering       │
│    (Apple II core + Canvas output)      │
└─────────────────────────────────────────┘
```

## Deployment

### Static Hosting

The built application is fully static and can be deployed anywhere:

1. **Build everything**:
   ```bash
   ./build-wasm.sh
   npm run build
   ```

2. **Upload the `dist/` directory** to any static hosting:
   - Netlify
   - Vercel
   - GitHub Pages
   - AWS S3 + CloudFront
   - Any web server

### Server Configuration

**MIME Types**

Ensure your server serves the correct MIME type for WASM:

```
.wasm -> application/wasm
```

**Headers for WASM**

Vite automatically configures these in development:
```
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
```

For production, configure your server/CDN similarly.

**Compression**

Enable gzip or brotli compression:

```nginx
# Nginx example
location ~ \.wasm$ {
    gzip_static on;
    add_header Content-Type application/wasm;
}
```

### Example: Netlify

1. Create `netlify.toml`:
   ```toml
   [build]
     command = "./build-wasm.sh && npm run build"
     publish = "dist"

   [[headers]]
     for = "/*"
     [headers.values]
       Cross-Origin-Embedder-Policy = "require-corp"
       Cross-Origin-Opener-Policy = "same-origin"

   [[headers]]
     for = "*.wasm"
     [headers.values]
       Content-Type = "application/wasm"
   ```

2. Connect your repository to Netlify
3. Deploy!

## Supported Disk Formats

- `.dsk` - DOS 3.3 order disk image
- `.woz` - WOZ disk image format
- `.nib` - Nibble format
- `.po` - ProDOS order disk image
- `.2mg` - 2IMG format (with header)
- `.zip` - Compressed disk images
- `.gz` - Gzipped disk images

## WASM API Reference

The Go code exports the following API to JavaScript:

```typescript
interface EmulatorAPI {
  // Core control
  reset(): void;
  pause(): void;
  resume(): void;

  // Disk operations
  loadDisk(drive: number, data: Uint8Array, filename: string): Promise<string | null>;
  loadDiskFromURL(drive: number, url: string): Promise<string | null>;

  // Input (for advanced use - keyboard is handled automatically)
  sendChar(char: number): void;
  sendText(text: string): void;

  // State queries
  isPaused(): boolean;
  getFrequency(): number;
  getDiskInfo(drive: number): DiskInfo | null;

  // Configuration
  toggleSpeed(): void;
  setScreenMode(mode: 'ntsc' | 'plain' | 'green'): void;

  // Screenshot
  screenshot(): void;
}

// Access via:
window.wasmAPI.reset();
window.wasmAPI.loadDisk(1, data, "game.dsk");
```

## Performance

- **WASM Size**: ~26 MB uncompressed, ~7 MB gzipped
- **Load Time**: 2-5 seconds on broadband
- **CPU Usage**: ~50-70% of one core on modern hardware
- **Emulation Speed**: 100% of native speed
- **Frame Rate**: 60 FPS (Ebiten rendering)

## Troubleshooting

### "Failed to load emulator"

- Check that WASM files exist in `public/wasm/`
- Run `./build-wasm.sh` to rebuild
- Check browser console for detailed errors
- Verify CORS headers are set correctly

### "Disk loading failed"

- Ensure disk image format is supported
- Check file is not corrupted
- For URL loading, verify CORS is enabled on the source

### No audio

- Some browsers require user interaction before playing audio
- Click the emulator screen to enable audio context
- Check browser console for audio errors

### Keys not working

- Click on the emulator canvas to focus it
- Keyboard is handled by Ebiten's native input system
- Ensure no browser extensions are intercepting keys

### Build errors

```bash
# Clean and rebuild
rm -rf node_modules package-lock.json
npm install
./build-wasm.sh
npm run build
```

## Development

### Project Scripts

```bash
# Development server with hot reload
npm run dev

# Build WASM binary
./build-wasm.sh

# Build React app for production
npm run build

# Type checking
npm run type-check

# Preview production build
npm run preview
```

### Modifying the UI

- Edit files in `src/` - Vite will auto-reload
- React components use Material-UI
- TypeScript ensures type safety
- WASM rebuild only needed for Go changes

### Modifying the Emulator

- Edit files in `go/`
- Run `./build-wasm.sh` to rebuild
- Refresh browser to load new WASM

## Known Limitations

1. **No virtual keyboard** - Mobile devices need external keyboard
2. **No save states** - Page refresh loses state
3. **Read-only disks** - Can't write to loaded disk images
4. **No gamepad** - Only keyboard supported

## Future Enhancements

- Virtual keyboard overlay for mobile devices
- Save/load state with IndexedDB
- PWA support (offline capability)
- Disk write support
- Multiple screen sizes/zoom levels
- More sample disk library

## Credits

- [izapple2](https://github.com/ivanizag/izapple2) by Ivan Izaguirre
- [Ebiten](https://ebiten.org/) game engine
- [React](https://react.dev/) UI library
- [Material-UI](https://mui.com/) component library
- [Vite](https://vitejs.dev/) build tool

## License

Same as the parent izapple2 project.
