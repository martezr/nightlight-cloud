// ipxeagent runs inside a Linux network namespace that is attached to an OVS
// bridge. It listens on the namespace interface address (169.254.169.253:8170)
// and reverse-proxies every request to the main nightlight-cloud iPXE server
// via a Unix domain socket, keeping the namespace's network stack completely
// separate from the host.
//
// Typical invocation (after ip netns exec or nsenter):
//
//	ipxeagent -listen 169.254.169.253:8170 -socket /opt/nightlight/ipxe.sock -log /var/log/ipxeagent.log
package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	listenAddr := flag.String("listen", "169.254.169.253:8170", "address to listen on inside the namespace")
	socketPath := flag.String("socket", "/opt/nightlight/ipxe.sock", "Unix socket path of the main iPXE server")
	logPath := flag.String("log", "/var/log/ipxeagent.log", "log file path (use - for stderr only)")
	flag.Parse()

	logger := buildLogger(*logPath)
	proxy := buildProxy(*socketPath, logger)

	srv := &http.Server{
		Addr:         *listenAddr,
		Handler:      loggingHandler(proxy, logger),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	go func() {
		logger.Printf("ipxeagent: listening on %s → unix:%s", *listenAddr, *socketPath)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("ipxeagent: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Printf("ipxeagent: shutdown error: %v", err)
	}
	logger.Printf("ipxeagent: stopped")
}

// buildProxy creates a reverse proxy that dials the main iPXE server over the
// given Unix socket. The request path and headers are forwarded verbatim.
func buildProxy(socketPath string, logger *log.Logger) *httputil.ReverseProxy {
	target, _ := url.Parse("http://ipxe-server")
	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Transport = &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
		// Keep connections alive across requests to avoid repeated socket dials.
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Printf("proxy error: %s %s: %v", r.Method, r.URL.RequestURI(), err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}

	return proxy
}

// loggingHandler wraps h to log every request's method, path, remote addr,
// status code, and elapsed time.
func loggingHandler(h http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		h.ServeHTTP(rw, r)
		logger.Printf("%s %s %s → %d (%s)",
			r.RemoteAddr, r.Method, r.URL.RequestURI(), rw.status, time.Since(start))
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// buildLogger opens logPath for appending and returns a logger that writes to
// both that file and stderr. If logPath is "-" or cannot be opened, stderr only.
func buildLogger(logPath string) *log.Logger {
	writers := []io.Writer{os.Stderr}
	if logPath != "" && logPath != "-" {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("ipxeagent: cannot open log file %s: %v; using stderr only", logPath, err)
		} else {
			writers = append(writers, f)
		}
	}
	return log.New(io.MultiWriter(writers...), "", log.LstdFlags)
}
