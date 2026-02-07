import {
  Box,
  IconButton,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Tooltip,
  SelectChangeEvent,
} from '@mui/material';
import {
  PlayArrow,
  Pause,
  Refresh,
  Speed,
  CameraAlt,
} from '@mui/icons-material';
import { emulator } from '../api/emulator';
import type { ScreenMode } from '../api/types';

interface ControlPanelProps {
  paused: boolean;
}

export function ControlPanel({ paused }: ControlPanelProps) {
  const handlePauseToggle = () => {
    if (paused) {
      emulator.resume();
    } else {
      emulator.pause();
    }
  };

  const handleReset = () => {
    if (window.confirm('Reset the emulator? This will clear memory.')) {
      emulator.reset();
    }
  };

  const handleScreenModeChange = (event: SelectChangeEvent) => {
    emulator.setScreenMode(event.target.value as ScreenMode);
  };

  const handleScreenshot = () => {
    emulator.screenshot();
  };

  const handleSpeedToggle = () => {
    emulator.toggleSpeed();
  };

  return (
    <Box
      sx={{
        display: 'flex',
        gap: 2,
        alignItems: 'center',
        padding: 2,
        backgroundColor: 'background.paper',
        borderBottom: 1,
        borderColor: 'divider',
      }}
    >
      <Tooltip title={paused ? 'Resume' : 'Pause'}>
        <IconButton color="primary" onClick={handlePauseToggle}>
          {paused ? <PlayArrow /> : <Pause />}
        </IconButton>
      </Tooltip>

      <Tooltip title="Reset">
        <IconButton color="secondary" onClick={handleReset}>
          <Refresh />
        </IconButton>
      </Tooltip>

      <Tooltip title="Toggle Speed (Full/NTSC)">
        <IconButton onClick={handleSpeedToggle}>
          <Speed />
        </IconButton>
      </Tooltip>

      <Tooltip title="Screenshot">
        <IconButton onClick={handleScreenshot}>
          <CameraAlt />
        </IconButton>
      </Tooltip>

      <FormControl size="small" sx={{ minWidth: 120 }}>
        <InputLabel>Screen Mode</InputLabel>
        <Select
          defaultValue="ntsc"
          label="Screen Mode"
          onChange={handleScreenModeChange}
        >
          <MenuItem value="ntsc">NTSC</MenuItem>
          <MenuItem value="plain">Plain</MenuItem>
          <MenuItem value="green">Green</MenuItem>
          <MenuItem value="amber">Amber</MenuItem>
        </Select>
      </FormControl>
    </Box>
  );
}
