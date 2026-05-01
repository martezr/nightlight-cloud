import { useEffect, useRef } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { AttachAddon } from '@xterm/addon-attach';
import '@xterm/xterm/css/xterm.css'; // Import the xterm.js CSS

const TerminalComponent = () => {
  const terminalRef = useRef(null);
  const term = useRef<any>(null);

  useEffect(() => {
    // 1. Instantiate the terminal
    term.current = new Terminal({
      cols: 120, // Set width in character columns
      rows: 24, // Set height in character rows
      // or use CSS to define pixel dimensions on the container
    });
    const fitAddon = new FitAddon();
    term.current.loadAddon(fitAddon);
    fitAddon.fit(); // Fit the terminal to the parent container initially

    // 2. Open the terminal in the DOM element
    term.current.open(terminalRef.current);
    term.current.write('Requesting a cloud shell...\r\n');

    // 3. Establish WebSocket connection
    // Replace 'ws://localhost:8080' with your backend WebSocket URL
    const socket = new WebSocket('ws://10.0.0.237/ssh'); 
    socket.binaryType = 'arraybuffer';

    // 4. Use the AttachAddon to bind the terminal and WebSocket
    const attachAddon = new AttachAddon(socket);
    term.current.loadAddon(attachAddon);

    // Optional: Handle connection status messages
    socket.onopen = () => {
      term.current.write('Connection established.\r\n\n');
    };

    //socket.onerror = (error) => {
    //  term.current.write(`WebSocket Error: ${error}\n`);
    //};

    // Handle window resize events
    const handleResize = () => {
      fitAddon.fit();
    };
    window.addEventListener('resize', handleResize);

  /*  // 5. Handle terminal resizing (requires backend implementation to handle 'resize' events)
    const handleResize = () => {
        const { rows, cols } = term.current.size;
        socket.send(JSON.stringify({ type: 'resize', rows, cols }));
    };
    // Note: A fit addon or custom logic is often needed for dynamic resizing
    window.addEventListener('resize', handleResize);
*/

    // Cleanup function
    return () => {
      window.removeEventListener('resize', handleResize);
      socket.close();
      term.current.dispose();
    };
  }, []);

  return <div ref={terminalRef} style={{ height: '100%', width: '100%'}} />;
};

export default TerminalComponent;