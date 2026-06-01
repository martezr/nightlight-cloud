import { useEffect, useRef } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { AttachAddon } from '@xterm/addon-attach';
import '@xterm/xterm/css/xterm.css';

const TerminalComponent = () => {
  // This div is placed inside `absolute inset-0` in app-sidebar, so it always
  // has explicit pixel dimensions that FitAddon's getComputedStyle() can read.
  const terminalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!terminalRef.current) return;

    const term = new Terminal({
      cursorBlink: true,
      fontFamily: 'Menlo, Monaco, "Cascadia Code", "Courier New", monospace',
      fontSize: 13,
    });
    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(terminalRef.current);

    // The container has explicit pixel dimensions (absolute positioning), so
    // a single rAF is enough for layout to stabilise before we fit.
    requestAnimationFrame(() => fitAddon.fit());

    term.write('Requesting a cloud shell...\r\n');

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const socket = new WebSocket(`${protocol}//${window.location.host}/ssh`);
    socket.binaryType = 'arraybuffer';

    const attachAddon = new AttachAddon(socket);
    term.loadAddon(attachAddon);

    socket.onopen = () => {
      term.write('Connection established.\r\n\n');
      fitAddon.fit();
    };

    // Refit whenever the container changes size (sidebar collapse, window resize)
    const ro = new ResizeObserver(() => fitAddon.fit());
    ro.observe(terminalRef.current!);

    const handleWindowResize = () => fitAddon.fit();
    window.addEventListener('resize', handleWindowResize);

    return () => {
      window.removeEventListener('resize', handleWindowResize);
      ro.disconnect();
      socket.close();
      term.dispose();
    };
  }, []);

  return <div ref={terminalRef} className="h-full w-full" />;
};

export default TerminalComponent;
