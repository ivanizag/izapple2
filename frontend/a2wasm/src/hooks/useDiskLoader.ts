import { useState } from 'react';
import { emulator } from '../api/emulator';

export function useDiskLoader() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadDiskFile = async (drive: number, file: File): Promise<boolean> => {
    setLoading(true);
    setError(null);

    try {
      const buffer = await file.arrayBuffer();
      const data = new Uint8Array(buffer);
      const result = await emulator.loadDisk(drive, data, file.name);

      if (result) {
        setError(result);
        return false;
      }

      return true;
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setError(message);
      return false;
    } finally {
      setLoading(false);
    }
  };

  const loadDiskFromURL = async (drive: number, url: string): Promise<boolean> => {
    setLoading(true);
    setError(null);

    try {
      const result = await emulator.loadDiskFromURL(drive, url);

      if (result) {
        setError(result);
        return false;
      }

      return true;
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setError(message);
      return false;
    } finally {
      setLoading(false);
    }
  };

  return {
    loading,
    error,
    loadDiskFile,
    loadDiskFromURL,
  };
}
