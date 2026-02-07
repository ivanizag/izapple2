import { useState, useEffect } from 'react';
import { loadWASM } from '../api/wasm-loader';
import type { DiskInfo } from '../api/types';

export interface EmulatorState {
  ready: boolean;
  loading: boolean;
  error: string | null;
  paused: boolean;
  frequency: number;
  drive1: DiskInfo | null;
  drive2: DiskInfo | null;
}

export function useEmulator() {
  const [state, setState] = useState<EmulatorState>({
    ready: false,
    loading: true,
    error: null,
    paused: false,
    frequency: 0,
    drive1: null,
    drive2: null,
  });

  useEffect(() => {
    // Initialize WASM
    loadWASM()
      .then(() => {
        setState(s => ({ ...s, ready: true, loading: false }));

        // Poll for state updates every 2 seconds
        const interval = setInterval(() => {
          if (window.wasmAPI) {
            setState(s => ({
              ...s,
              paused: window.wasmAPI!.isPaused(),
              frequency: window.wasmAPI!.getFrequency(),
              drive1: window.wasmAPI!.getDiskInfo(0),
              drive2: window.wasmAPI!.getDiskInfo(1),
            }));
          }
        }, 2000);

        return () => clearInterval(interval);
      })
      .catch((error: Error) => {
        setState(s => ({
          ...s,
          loading: false,
          error: error.message,
        }));
      });
  }, []);

  return state;
}
