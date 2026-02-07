// izapple2 WebAssembly Loader and JavaScript Bridge

let wasmReady = false;
let currentDrive = 1;

// Initialize WASM when page loads
window.addEventListener('load', async () => {
    console.log('Initializing izapple2 WebAssembly...');
    await initWASM();
});

// Initialize the WebAssembly module
async function initWASM() {
    const go = new Go();

    updateLoadingStatus('Downloading WASM module...');

    try {
        const result = await WebAssembly.instantiateStreaming(
            fetch('izapple2.wasm'),
            go.importObject
        );

        updateLoadingStatus('Starting emulator...');

        // Run the Go WASM module
        go.run(result.instance);

        // Wait a moment for initialization
        setTimeout(() => {
            wasmReady = true;
            hideLoading();
            initUI();
            console.log('izapple2 WebAssembly ready!');

            // Check for URL parameters to load disks
            checkURLParameters();
        }, 500);

    } catch (err) {
        console.error('Failed to load WASM:', err);
        updateLoadingStatus('Error loading emulator: ' + err.message);
    }
}

// Update loading screen status
function updateLoadingStatus(message) {
    const statusEl = document.getElementById('loading-status');
    if (statusEl) {
        statusEl.textContent = message;
    }
}

// Hide loading screen and show app
function hideLoading() {
    document.getElementById('loading-screen').style.display = 'none';
    document.getElementById('app-container').style.display = 'block';
}

// Initialize UI event handlers
function initUI() {
    // Disk loading buttons
    document.getElementById('load-disk-1').addEventListener('click', () => {
        currentDrive = 1;
        document.getElementById('file-input').click();
    });

    document.getElementById('load-disk-2').addEventListener('click', () => {
        currentDrive = 2;
        document.getElementById('file-input').click();
    });

    // File input change handler
    document.getElementById('file-input').addEventListener('change', handleFileSelect);

    // Control buttons
    document.getElementById('reset-btn').addEventListener('click', () => {
        if (wasmReady && window.resetEmulator) {
            window.resetEmulator();
        }
    });

    document.getElementById('pause-btn').addEventListener('click', () => {
        if (wasmReady && window.togglePause) {
            window.togglePause();
            // Note: We'd need to track pause state for UI update
        }
    });

    document.getElementById('screenshot-btn').addEventListener('click', () => {
        if (wasmReady && window.downloadScreenshot) {
            const result = window.downloadScreenshot();
            if (result) {
                alert('Screenshot error: ' + result);
            }
        }
    });

    // Sample disk buttons
    document.querySelectorAll('.sample-disk').forEach(btn => {
        btn.addEventListener('click', () => {
            const url = btn.dataset.url;
            const drive = parseInt(btn.dataset.drive);
            loadDiskFromURL(drive, url);
        });
    });

    // Drag and drop setup
    setupDragAndDrop();
}

// Handle file selection from file picker
async function handleFileSelect(event) {
    const file = event.target.files[0];
    if (!file) return;

    await loadDiskFile(currentDrive, file);

    // Reset file input
    event.target.value = '';
}

// Load a disk file into the specified drive
async function loadDiskFile(drive, file) {
    if (!wasmReady) {
        alert('Emulator not ready yet');
        return;
    }

    console.log(`Loading ${file.name} into drive ${drive}...`);
    updateDiskStatus(drive, `Loading ${file.name}...`);

    try {
        const buffer = await file.arrayBuffer();
        const bytes = new Uint8Array(buffer);

        if (window.loadDisk) {
            const result = window.loadDisk(drive, bytes, file.name);
            if (result) {
                alert('Error loading disk: ' + result);
                updateDiskStatus(drive, 'Empty');
            } else {
                updateDiskStatus(drive, file.name);
                console.log(`Successfully loaded disk ${drive}: ${file.name}`);
            }
        } else {
            alert('Error: loadDisk function not available');
            updateDiskStatus(drive, 'Empty');
        }
    } catch (err) {
        console.error('Error reading file:', err);
        alert('Error reading file: ' + err.message);
        updateDiskStatus(drive, 'Empty');
    }
}

// Load a disk from URL
function loadDiskFromURL(drive, url) {
    if (!wasmReady) {
        alert('Emulator not ready yet');
        return;
    }

    console.log(`Loading disk from URL into drive ${drive}: ${url}`);
    updateDiskStatus(drive, 'Loading from URL...');

    if (window.loadDiskFromURL) {
        const result = window.loadDiskFromURL(drive, url);
        if (result) {
            alert('Error loading disk: ' + result);
            updateDiskStatus(drive, 'Empty');
        } else {
            const filename = url.substring(url.lastIndexOf('/') + 1);
            updateDiskStatus(drive, filename);
            console.log(`Successfully loaded disk ${drive} from URL`);
        }
    } else {
        alert('Error: loadDiskFromURL function not available');
        updateDiskStatus(drive, 'Empty');
    }
}

// Update disk status display
function updateDiskStatus(drive, status) {
    const statusEl = document.getElementById(`disk-${drive}-status`);
    if (statusEl) {
        statusEl.textContent = `Drive ${drive}: ${status}`;
    }
}

// Setup drag and drop functionality
function setupDragAndDrop() {
    const dropZone = document.getElementById('drop-zone');
    const emulatorContainer = document.getElementById('emulator-canvas-container');

    // Prevent default drag behaviors on the whole page
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        document.body.addEventListener(eventName, preventDefaults, false);
    });

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    // Highlight drop zone when dragging over
    ['dragenter', 'dragover'].forEach(eventName => {
        emulatorContainer.addEventListener(eventName, () => {
            dropZone.classList.add('drag-over');
        }, false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        emulatorContainer.addEventListener(eventName, () => {
            dropZone.classList.remove('drag-over');
        }, false);
    });

    // Handle dropped files
    emulatorContainer.addEventListener('drop', handleDrop, false);

    async function handleDrop(e) {
        const dt = e.dataTransfer;
        const files = dt.files;

        if (files.length > 0) {
            // Load first file into drive 1, second into drive 2
            for (let i = 0; i < Math.min(files.length, 2); i++) {
                await loadDiskFile(i + 1, files[i]);
            }
        }
    }
}

// Check URL parameters for disk loading
function checkURLParameters() {
    const params = new URLSearchParams(window.location.search);

    const disk1 = params.get('disk1');
    if (disk1) {
        console.log('Loading disk 1 from URL parameter:', disk1);
        loadDiskFromURL(1, disk1);
    }

    const disk2 = params.get('disk2');
    if (disk2) {
        console.log('Loading disk 2 from URL parameter:', disk2);
        loadDiskFromURL(2, disk2);
    }
}

// Trigger download of a file (called from Go)
window.triggerDownload = function(data, filename) {
    const blob = new Blob([data], { type: 'application/octet-stream' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    console.log(`Downloaded: ${filename}`);
};

// FPS counter (optional)
setInterval(() => {
    const fpsEl = document.getElementById('fps-counter');
    if (fpsEl && wasmReady) {
        // Note: Actual FPS would need to be exposed from Go
        // For now, just show that we're running
        fpsEl.textContent = 'Running';
    }
}, 1000);

// Log when WASM functions become available
const checkFunctions = setInterval(() => {
    if (window.loadDisk && window.loadDiskFromURL) {
        console.log('WASM bridge functions available');
        clearInterval(checkFunctions);
    }
}, 100);
