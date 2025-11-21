package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	fmt.Println("Hello, World!")

	// add golang chi router
	r := chi.NewRouter()
	fmt.Println(r)

	// add middleware to print request ip address
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Request from:", r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/", getAPIVersionsHandler)
	r.Put("/latest/api/token", fetchAPITokenHandler)
	r.Get("/{version}/meta-data/", getMetaDataHandler)
	r.Get("/{version}/user-data/", getMetaDataHandler)
	r.Get("/{version}/meta-data/instance-id", getMetaDataHandler)
	r.Get("/{version}/meta-data/mac", getMetaDataHandler)

	// start server on port 80
	http.ListenAndServe(":80", r)
}

func fetchAPITokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func getAPIVersionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("latest"))
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
