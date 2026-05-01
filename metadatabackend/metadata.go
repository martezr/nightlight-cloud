package metadatabackend

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"log"

	"github.com/asdine/storm/v3"
	"github.com/go-chi/chi/v5"
	"github.com/martezr/nightlight-cloud/utils"
)

var (
	localdb *storm.DB
)

// StartMetadataServer starts a chi-based HTTP server listening on a UNIX socket.
// Call with ctx to allow graceful shutdown. If sockPath is empty, "/opt/nightlight/metadata.sock" is used.
func StartMetadataServer(ctx context.Context, db *storm.DB) error {
	sockPath := "/opt/nightlight/metadata.sock"

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

	r.Get("/latest/meta-data/instance-id", getMetaDataHandler)
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

func getMetaDataHandler(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, you would query the database for the instance metadata
	// For this example, we'll just return a static instance ID
	xff := r.Header.Get("X-Forwarded-For")
	clientIP := ""
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			clientIP = strings.TrimSpace(parts[0])
			if net.ParseIP(clientIP) == nil {
				clientIP = ""
			}
		}
	}
	if clientIP == "" {
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			clientIP = host
		} else {
			clientIP = r.RemoteAddr
		}
	}

	log.Printf("Received metadata request from %s for %s", clientIP, r.URL.Path)
	var instance utils.Instance
	err := localdb.One("MetadataIPAddress", clientIP, &instance)
	if err != nil {
		log.Printf("Error retrieving instance metadata: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error retrieving metadata"))
		return
	}
	w.Write([]byte(instance.ID))
}
