import { useRef, useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Typography,
  CircularProgress,
  Alert,
} from '@mui/material';
import { Storage, CloudUpload } from '@mui/icons-material';
import { useDiskLoader } from '../hooks/useDiskLoader';
import type { DiskInfo } from '../api/types';

interface DiskSlotProps {
  drive: number;
  diskInfo: DiskInfo | null;
  onLoad: (file: File) => void;
  loading: boolean;
}

function DiskSlot({ drive, diskInfo, onLoad, loading }: DiskSlotProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      onLoad(file);
    }
    // Reset input
    e.target.value = '';
  };

  return (
    <Card variant="outlined">
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Storage />
          <Box sx={{ flex: 1 }}>
            <Typography variant="subtitle2" color="text.secondary">
              Drive {drive}
            </Typography>
            <Typography variant="body2">
              {loading ? (
                <CircularProgress size={16} />
              ) : diskInfo?.loaded ? (
                diskInfo.name
              ) : (
                'Empty'
              )}
            </Typography>
          </Box>
          <Button
            variant="outlined"
            size="small"
            onClick={handleClick}
            disabled={loading}
          >
            Load
          </Button>
          <input
            ref={fileInputRef}
            type="file"
            accept=".dsk,.woz,.nib,.po,.2mg,.zip,.gz"
            style={{ display: 'none' }}
            onChange={handleFileChange}
          />
        </Box>
      </CardContent>
    </Card>
  );
}

interface DiskManagerProps {
  drive1: DiskInfo | null;
  drive2: DiskInfo | null;
}

export function DiskManager({ drive1, drive2 }: DiskManagerProps) {
  const { loading, error, loadDiskFile, loadDiskFromURL } = useDiskLoader();
  const [dragOver, setDragOver] = useState(false);

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);

    const files = Array.from(e.dataTransfer.files);
    if (files.length > 0) {
      await loadDiskFile(1, files[0]);
    }
    if (files.length > 1) {
      await loadDiskFile(2, files[1]);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  };

  const handleDragLeave = () => {
    setDragOver(false);
  };

  const handleLoadSampleDisk = async () => {
    await loadDiskFromURL(1, '<internal>/dos33.dsk');
  };

  return (
    <Box
      sx={{ padding: 2, minWidth: 300 }}
      onDrop={handleDrop}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
    >
      <Typography variant="h6" gutterBottom>
        Disk Drives
      </Typography>

      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mb: 2 }}>
        <DiskSlot
          drive={1}
          diskInfo={drive1}
          onLoad={(file) => loadDiskFile(1, file)}
          loading={loading}
        />
        <DiskSlot
          drive={2}
          diskInfo={drive2}
          onLoad={(file) => loadDiskFile(2, file)}
          loading={loading}
        />
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {dragOver && (
        <Box
          sx={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: 'rgba(0, 255, 0, 0.1)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000,
            pointerEvents: 'none',
          }}
        >
          <Box
            sx={{
              textAlign: 'center',
              color: 'primary.main',
            }}
          >
            <CloudUpload sx={{ fontSize: 64 }} />
            <Typography variant="h5">Drop disk image here</Typography>
          </Box>
        </Box>
      )}

      <Typography variant="h6" gutterBottom sx={{ mt: 3 }}>
        Sample Disks
      </Typography>
      <Button
        variant="outlined"
        fullWidth
        onClick={handleLoadSampleDisk}
        disabled={loading}
      >
        Load DOS 3.3
      </Button>
    </Box>
  );
}
