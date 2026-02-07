// TypeScript types for the Go WASM API

export interface EmulatorAPI {
  // Core control
  reset: () => void;
  pause: () => void;
  resume: () => void;

  // Disk operations
  loadDisk: (drive: number, data: Uint8Array, filename: string) => Promise<string | null>;
  loadDiskFromURL: (drive: number, url: string) => Promise<string | null>;

  // Input
  sendKey: (keyCode: number) => void;
  sendText: (text: string) => void;

  // State queries
  isPaused: () => boolean;
  getFrequency: () => number;
  getDiskInfo: (drive: number) => DiskInfo | null;

  // Configuration
  toggleSpeed: () => void;
  setScreenMode: (mode: ScreenMode) => void;

  // Screenshot
  screenshot: () => void;
}

export interface DiskInfo {
  name: string;
  loaded: boolean;
}

export type ScreenMode = 'ntsc' | 'plain' | 'green' | 'amber';

// Extend Window interface for WASM API
declare global {
  interface Window {
    wasmAPI?: EmulatorAPI;
    Go?: any;
    triggerDownload?: (data: Uint8Array, filename: string) => void;
  }
}
