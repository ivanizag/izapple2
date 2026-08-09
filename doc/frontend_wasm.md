# The WebAssembly frontend

`a2wasm` is the emulator compiled to WebAssembly, running in the browser. The
machine and the screen are the Go code of [a2ebiten](frontend_ebiten.md) built
for `js/wasm`, and around it there is a React interface with the controls and
the disk drives.

There is nothing to install for whoever opens the page, and there is a fair
amount to build: it is the only frontend that needs Node besides Go.

## Building

You need Go and Node 18 or newer.

``` terminal
git clone github.com/ivanizag/izapple2
cd izapple2/frontend/a2wasm
npm install
npm run build
```

`npm run build` compiles the Go code to WebAssembly, type checks the React
code and puts everything in `dist`, which is a static site that any web server
can serve. To build only the WebAssembly part, run `./build-wasm.sh`: it leaves
`izapple2.wasm`, its gzipped copy and the `wasm_exec.js` of your Go
installation in `public/wasm`.

The binary is around 26 MB, 7 MB gzipped, so serve it compressed.

## Running it while working on it

``` terminal
casa@servidor:~$ npm run dev
```

Then open <http://localhost:5173>. Vite reloads the page when the React code
changes; a change in the Go code needs `./build-wasm.sh` and a refresh.

## Serving it

`dist` is a static site, with two things to get right in the server:

- `.wasm` files have to be served as `application/wasm`.
- The page asks for `Cross-Origin-Embedder-Policy: require-corp` and
  `Cross-Origin-Opener-Policy: same-origin`, which Vite sets by itself while
  developing.

## Using it

The machine boots by itself when the page loads. The toolbar on top has pause
and resume, reset, a screenshot that downloads a PNG, and the screen mode: NTSC
colour, plain or green. The bar at the bottom shows the speed in MHz.

The keyboard goes to the emulator once you click on the screen. There are no
joysticks and no mouse.

To insert a diskette, use the file picker of a drive or drop the file on the
disk panel: the first one goes to drive 1 and the second to drive 2. There is
also a button to load DOS 3.3 from the resources embedded in the binary. The
formats are the same as in the other frontends, compressed ones included.

## Driving it from JavaScript

The Go code exports `window.wasmAPI`, which is what the React interface uses
and is there for whatever else you want to do with it:

``` javascript
wasmAPI.reset()
wasmAPI.pause()
wasmAPI.resume()
wasmAPI.isPaused()
wasmAPI.getFrequency()
wasmAPI.setScreenMode("ntsc")   // "ntsc", "plain" or "green"
wasmAPI.screenshot()
wasmAPI.sendKey(13)
wasmAPI.sendText("10 PRINT \"HELLO\"\n")
wasmAPI.loadDisk(1, bytes, "game.dsk")
wasmAPI.loadDiskFromURL(1, "https://example.com/game.dsk")
wasmAPI.loadDiskFromURL(1, "<internal>/dos33.dsk")
```

An URL is fetched by the browser, so it has to allow it with CORS.

## What is missing

- **Command line options.** The machine is the default one and there is no way
  to choose a model or to configure the slots yet.
- **Joysticks and mouse.**
- **Sound in some browsers** until you click on the page: they do not start
  audio before the visitor asks for something.
- **State that survives a refresh.** Reloading the page starts a new machine.
- **`getDiskInfo`**, which is in the API and always answers null, so the drives
  do not show what is in them.

The notes on how it is put together, the file layout and the deployment
recipes are in
[frontend/a2wasm/README.md](../frontend/a2wasm/README.md).
