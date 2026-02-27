import { Box, CircularProgress, Alert, Typography } from '@mui/material';
import { ThemeProvider, CssBaseline } from '@mui/material';
import { theme } from './styles/theme';
import { useEmulator } from './hooks/useEmulator';
import { EmulatorScreen } from './components/EmulatorScreen';
import { ControlPanel } from './components/ControlPanel';
import { DiskManager } from './components/DiskManager';
import { StatusBar } from './components/StatusBar';

function App() {
  const emulatorState = useEmulator();

  if (emulatorState.loading) {
    return (
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            height: '100vh',
            gap: 2,
          }}
        >
          <CircularProgress size={60} />
          <Typography variant="h5" color="primary">
            Loading izapple2 Emulator...
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Initializing WebAssembly
          </Typography>
        </Box>
      </ThemeProvider>
    );
  }

  if (emulatorState.error) {
    return (
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            height: '100vh',
            padding: 4,
          }}
        >
          <Alert severity="error">
            <Typography variant="h6">Failed to load emulator</Typography>
            <Typography variant="body2">{emulatorState.error}</Typography>
          </Alert>
        </Box>
      </ThemeProvider>
    );
  }

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          height: '100vh',
          overflow: 'hidden',
        }}
      >
        <ControlPanel paused={emulatorState.paused} />

        <Box sx={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
          <DiskManager
            drive1={emulatorState.drive1}
            drive2={emulatorState.drive2}
          />
          <EmulatorScreen />
        </Box>

        <StatusBar
          frequency={emulatorState.frequency}
          paused={emulatorState.paused}
        />
      </Box>
    </ThemeProvider>
  );
}

export default App;
