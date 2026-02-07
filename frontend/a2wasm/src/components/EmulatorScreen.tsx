import { useEffect, useRef } from 'react';
import { Box } from '@mui/material';

export function EmulatorScreen() {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Move Ebiten canvas into our container
    const checkCanvas = setInterval(() => {
      // Look for canvas in body (where Ebiten creates it)
      const canvas = document.querySelector('body > canvas');
      if (canvas && containerRef.current) {
        console.log('Found Ebiten canvas, moving to container');

        // Remove canvas from body
        if (canvas.parentNode) {
          canvas.parentNode.removeChild(canvas);
        }

        // Add to our container
        containerRef.current.appendChild(canvas);

        // Aggressively override ALL Ebiten canvas styles
        canvas.style.cssText = `
          position: static !important;
          top: auto !important;
          left: auto !important;
          right: auto !important;
          bottom: auto !important;
          max-width: 100%;
          max-height: 100%;
          width: auto;
          height: auto;
          display: block;
          margin: auto;
          image-rendering: pixelated;
        `;

        // Make canvas focusable and focus it
        canvas.setAttribute('tabindex', '0');
        canvas.focus();
        console.log('Canvas focused');

        clearInterval(checkCanvas);
        console.log('Canvas moved and styled successfully');
      }
    }, 100);

    // Cleanup after 10 seconds if canvas not found
    setTimeout(() => clearInterval(checkCanvas), 10000);

    return () => clearInterval(checkCanvas);
  }, []);

  return (
    <Box
      ref={containerRef}
      sx={{
        flex: 1,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: '#000',
        borderRadius: 1,
        overflow: 'hidden',
        border: '2px solid',
        borderColor: 'divider',
      }}
    />
  );
}
