package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
)

var (
	logFile *os.File
)

func main() {
	// Open log file for writing
	logFile, err := os.OpenFile("/var/log/metadataagent.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Create a logger instance with log file output
	logger := httplog.NewLogger("metadataagent", httplog.Options{
		LogLevel: slog.LevelDebug,
		JSON:     false, // set to true for production
		Concise:  true,  // concise format for development
		Writer:   logFile,
	})

	// add golang chi router
	r := chi.NewRouter()
	// 2. Use the RequestLogger middleware
	r.Use(httplog.RequestLogger(logger))

	// add middleware to print request ip address
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Request from:", r.RemoteAddr)
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
				r.Header.Set("X-Forwarded-For", prior+", "+host)
			} else {
				r.Header.Set("X-Forwarded-For", host)
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/", getAPIVersionsHandler)
	r.Put("/latest/api/token", fetchAPITokenHandler)
	r.Get("/{version}/meta-data/", getMetaDataHandler)
	r.Get("/{version}/user-data/", getMetaDataHandler)
	r.Get("/{version}/meta-data/instance-id", getInstanceIDHandler)
	r.Get("/{version}/meta-data/mac", getMetaDataHandler)

	r.Get("/{version}/meta-data/ami-id", getAMIIDHandler)
	r.Get("/{version}/meta-data/ami-launch-index", getMetaDataHandler)
	r.Get("/{version}/meta-data/ami-manifest-path", getMetaDataHandler)
	r.Get("/{version}/meta-data/block-device-mapping/", getMetaDataHandler)
	r.Get("/{version}/meta-data/events/", getMetaDataHandler)
	r.Get("/{version}/meta-data/hostname", getMetaDataHandler)
	r.Get("/{version}/meta-data/iam/", getMetaDataHandler)
	r.Get("/{version}/meta-data/instance-action", getMetaDataHandler)
	r.Get("/{version}/meta-data/instance-life-cycle", getMetaDataHandler)
	r.Get("/{version}/meta-data/instance-type", getMetaDataHandler)
	r.Get("/{version}/meta-data/local-hostname", getMetaDataHandler)
	r.Get("/{version}/meta-data/local-ipv4", getMetaDataHandler)
	r.Get("/{version}/meta-data/metrics/", getMetaDataHandler)
	r.Get("/{version}/meta-data/network/", getMetaDataHandler)
	r.Get("/{version}/meta-data/network/interfaces/macs/{mac}/", getMetaDataHandler)
	r.Get("/{version}/meta-data/placement/", getMetaDataHandler)
	r.Get("/{version}/meta-data/profile", getMetaDataHandler)
	r.Get("/{version}/meta-data/public-hostname", getMetaDataHandler)
	r.Get("/{version}/meta-data/public-ipv4", getMetaDataHandler)
	r.Get("/{version}/meta-data/public-keys/", getMetaDataHandler)
	r.Get("/{version}/meta-data/reservation-id", getMetaDataHandler)
	r.Get("/{version}/meta-data/security-groups", getMetaDataHandler)
	r.Get("/{version}/meta-data/services/", getMetaDataHandler)
	r.Get("/{version}/meta-data/tags/", getMetaDataHandler)

	// catch-all for deeper metadata paths
	r.Get("/{version}/meta-data/{rest:.*}", getMetaDataHandler)

	fmt.Println("Starting metadata service on 169.254.169.254:80")
	// start server on port 80
	err = http.ListenAndServe("169.254.169.254:80", r)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func fetchAPITokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func getAPIVersionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("latest"))
}

func getAMIIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ami-12345678"))
}

func getMetaDataHandler(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	fmt.Println("Version:", version)
	metadata := `ami-id
ami-launch-index
ami-manifest-path
block-device-mapping/
events/
hostname
iam/
instance-action
instance-id
instance-life-cycle
instance-type
local-hostname
local-ipv4
mac
metrics/
network/
placement/
profile
public-hostname
public-ipv4
public-keys/
reservation-id
security-groups
services/
tags/`

	w.Write([]byte(metadata))
}

func getInstanceIDHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
		clientIP = prior
	}
	host, _, err := net.SplitHostPort(clientIP)
	if err == nil {
		clientIP = host
	}

	log.SetOutput(logFile)
	log.Printf("Received request for instance ID from %s\n", clientIP)
	output := PlainTextUnixClient("/" + chi.URLParam(r, "version") + "/meta-data/instance-id")
	w.Write([]byte(output))
}

func PlainTextUnixClient(path string) string {
	socketPath := "/opt/nightlight/metadata.sock"

	// Create a custom DialContext for Unix Domain Sockets
	udsDialer := func(ctx context.Context, _, _ string) (net.Conn, error) {
		return net.Dial("unix", socketPath)
	}

	// Create a custom Transport with the UDS dialer
	transport := &http.Transport{
		DialContext: udsDialer,
	}

	// Create an HTTP client using the custom Transport
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second, // Optional: set a timeout
	}

	// Make a request to the UDS HTTP server
	resp, err := client.Get("http://unix-socket" + path) // The host name here is arbitrary, as it's ignored
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	fmt.Printf("Response Status: %s\n", resp.Status)

	// Read and print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return ""
	}
	fmt.Printf("Response Body: %s\n", string(body))
	return string(body)
}
