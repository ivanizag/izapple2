import { useEffect } from 'react';
import { emulator } from '../api/emulator';

// Apple II key code mappings
const keyMap: Record<string, number> = {
  'Escape': 27,
  'Backspace': 8,
  'Enter': 13,
  'ArrowLeft': 8,
  'ArrowRight': 21,
  'ArrowUp': 11,
  'ArrowDown': 10,
  'Tab': 9,
  'Delete': 127,
};

function mapKeyToApple2(e: KeyboardEvent): number | null {
  // Handle special keys
  if (keyMap[e.key]) {
    return keyMap[e.key];
  }

  // Handle Ctrl+key combinations
  if (e.ctrlKey && e.key.length === 1) {
    const char = e.key.toUpperCase();
    if (char >= 'A' && char <= 'Z') {
      return char.charCodeAt(0) - 64; // Ctrl+A = 1, etc.
    }
  }

  // Handle regular printable characters
  if (e.key.length === 1) {
    return e.key.toUpperCase().charCodeAt(0);
  }

  return null;
}

export function KeyboardHandler() {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't prevent default - let Ebiten also process the event

      // For printable characters, send as text (sendText will handle case via PutText/PutRune)
      if (e.key.length === 1 && !e.ctrlKey && !e.metaKey && !e.altKey) {
        try {
          emulator.sendText(e.key);
          // e.preventDefault(); // REMOVED - let event propagate to Ebiten canvas
        } catch (error) {
          console.error('Error sending text:', error);
        }
        return;
      }

      // For special keys, use sendKey with mapped codes
      const keyCode = mapKeyToApple2(e);
      if (keyCode !== null) {
        try {
          emulator.sendKey(keyCode);
          // e.preventDefault(); // REMOVED - let event propagate to Ebiten canvas
        } catch (error) {
          console.error('Error sending key:', error);
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  return null; // Invisible component
}
