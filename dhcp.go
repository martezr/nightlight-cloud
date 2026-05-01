package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"log"

	"github.com/asdine/storm/v3"
	"github.com/go-chi/chi/v5"
)

var (
	localdb *storm.DB
)

// StartDHCPServer starts a chi-based HTTP server listening on a UNIX socket.
// Call with ctx to allow graceful shutdown. If sockPath is empty, "/opt/nightlight/dhcp.sock" is used.
func StartDHCPServer(ctx context.Context, db *storm.DB) error {
	f, err := os.OpenFile("/var/log/dhcpbackend.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
	} else {
		mw := io.MultiWriter(os.Stdout, f)
		log.SetOutput(mw)
		defer f.Close()
	}

	sockPath := "/opt/nightlight/dhcp.sock"

	localdb = db
	// Remove any existing socket file
	if err := os.RemoveAll(sockPath); err != nil {
		return err
	}

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return err
	}
	// ensure socket file has reasonable permissions
	_ = os.Chmod(sockPath, 0660)

	r := chi.NewRouter()

	r.Get("/{mac}", func(w http.ResponseWriter, r *http.Request) {
		// Extract the {mac} variable
		mac := chi.URLParam(r, "mac")
		log.Printf("Received DHCP request for MAC: %s", mac)
		// Add colons back to the mac address if they were removed
		restoredMac := mac[:2] + ":" + mac[2:4] + ":" + mac[4:6] + ":" + mac[6:8] + ":" + mac[8:10] + ":" + mac[10:12]

		payload, found := FindInstanceByMacAddress(restoredMac)
		if found {
			log.Printf("Found DHCP payload for MAC: %s - %+v", restoredMac, payload)
		} else {
			log.Printf("No DHCP payload found for MAC: %s", restoredMac)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(payload)
	})
	// Example handlers
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("metadata service\n"))
	})

	srv := &http.Server{
		Handler: r,
	}

	errCh := make(chan error, 1)
	go func() {
		if serveErr := srv.Serve(ln); serveErr != nil && serveErr != http.ErrServerClosed {
			errCh <- serveErr
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		// graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("metadata server shutdown error: %v", err)
		}
		_ = ln.Close()
		_ = os.Remove(sockPath)
		return nil
	case err := <-errCh:
		_ = ln.Close()
		_ = os.Remove(sockPath)
		return err
	}
}
