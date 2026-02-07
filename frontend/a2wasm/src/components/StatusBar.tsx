import { Box, Typography, Chip } from '@mui/material';

interface StatusBarProps {
  frequency: number;
  paused: boolean;
}

export function StatusBar({ frequency, paused }: StatusBarProps) {
  return (
    <Box
      sx={{
        display: 'flex',
        gap: 2,
        alignItems: 'center',
        padding: 1,
        backgroundColor: 'background.paper',
        borderTop: 1,
        borderColor: 'divider',
      }}
    >
      <Typography variant="body2" color="text.secondary">
        izapple2 - Apple II Emulator
      </Typography>

      <Box sx={{ flex: 1 }} />

      {paused && (
        <Chip label="PAUSED" color="warning" size="small" />
      )}

      <Typography variant="body2" color="text.secondary">
        {frequency.toFixed(2)} MHz
      </Typography>

      <Typography variant="caption" color="text.secondary">
        Press F1 in emulator for help
      </Typography>
    </Box>
  );
}
