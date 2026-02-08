// WASM loader - initializes the Go WebAssembly module

export async function loadWASM(): Promise<void> {
  return new Promise((resolve, reject) => {
    // Load wasm_exec.js if not already loaded
    if (!window.Go) {
      const script = document.createElement('script');
      script.src = '/wasm/wasm_exec.js';
      script.onload = () => initWASM(resolve, reject);
      script.onerror = () => reject(new Error('Failed to load wasm_exec.js'));
      document.head.appendChild(script);
    } else {
      initWASM(resolve, reject);
    }
  });
}

async function initWASM(
  resolve: () => void,
  reject: (error: Error) => void
): Promise<void> {
  try {
    const go = new window.Go!();

    const result = await WebAssembly.instantiateStreaming(
      fetch('/wasm/izapple2.wasm'),
      go.importObject
    );

    // Run the Go WASM module (this starts the emulator)
    go.run(result.instance);

    // Wait a bit for the emulator to initialize and export functions
    setTimeout(() => {
      if (window.wasmAPI) {
        resolve();
      } else {
        console.error('WASM API not found on window object');
        reject(new Error('WASM API not initialized'));
      }
    }, 1000);
  } catch (error) {
    console.error('Failed to initialize WASM:', error);
    reject(error as Error);
  }
}
