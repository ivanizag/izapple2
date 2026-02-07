// Emulator API client - wraps the WASM API with error handling

import type { EmulatorAPI, DiskInfo, ScreenMode } from './types';

class EmulatorClient {
  private get api(): EmulatorAPI {
    if (!window.wasmAPI) {
      throw new Error('WASM API not initialized');
    }
    return window.wasmAPI;
  }

  reset(): void {
    this.api.reset();
  }

  pause(): void {
    this.api.pause();
  }

  resume(): void {
    this.api.resume();
  }

  async loadDisk(drive: number, data: Uint8Array, filename: string): Promise<string | null> {
    try {
      return await this.api.loadDisk(drive, data, filename);
    } catch (error) {
      console.error('Failed to load disk:', error);
      return String(error);
    }
  }

  async loadDiskFromURL(drive: number, url: string): Promise<string | null> {
    try {
      return await this.api.loadDiskFromURL(drive, url);
    } catch (error) {
      console.error('Failed to load disk from URL:', error);
      return String(error);
    }
  }

  sendKey(keyCode: number): void {
    this.api.sendKey(keyCode);
  }

  sendText(text: string): void {
    this.api.sendText(text);
  }

  isPaused(): boolean {
    return this.api.isPaused();
  }

  getFrequency(): number {
    return this.api.getFrequency();
  }

  getDiskInfo(drive: number): DiskInfo | null {
    return this.api.getDiskInfo(drive);
  }

  toggleSpeed(): void {
    this.api.toggleSpeed();
  }

  setScreenMode(mode: ScreenMode): void {
    this.api.setScreenMode(mode);
  }

  screenshot(): void {
    this.api.screenshot();
  }
}

export const emulator = new EmulatorClient();
