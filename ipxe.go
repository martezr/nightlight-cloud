package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/martezr/nightlight-cloud/utils"
)

// ipxeFilesDir is the root directory from which boot files are served.
// Operators place iPXE binaries, kernels, and initrds here.
const ipxeFilesDir = "/opt/nightlight/ipxe/files"

// ipxeListenAddr is the address the iPXE file server binds to.
// It must be reachable by VMs on the virtual network (not loopback).
const ipxeListenAddr = "0.0.0.0:8170"

// ipxeSocketPath is the Unix domain socket that ipxeagent processes running
// inside network namespaces use to reach this server.
const ipxeSocketPath = "/opt/nightlight/ipxe.sock"

// StartIPXEServer starts an HTTP server that provides two services to
// PXE-booting VMs:
//
//  1. GET /files/...   — static file tree rooted at ipxeFilesDir.
//     Place iPXE binaries (ipxe.efi, undionly.kpxe), kernels, and initrds
//     here.  The SDN controller's DHCP option 67 can point to files under
//     this prefix (e.g. http://<host>:8170/files/ipxe.efi).
//
//  2. GET /ipxe/{id}     — kickstart/preseed content for instance {id}.
//
//  3. GET /ipxe/mac/{mac} — same lookup but by NIC MAC address.
//     These endpoints serve the raw Kickstart field of the matching instance,
//     enabling iPXE scripts to reference
//     http://<host>:8170/ipxe/<instance-id> as an inst.ks= kernel param.
func StartIPXEServer(ctx context.Context, db *storm.DB) error {
	if err := os.MkdirAll(ipxeFilesDir, 0755); err != nil {
		log.Printf("iPXE server: cannot create files dir %s: %v", ipxeFilesDir, err)
	}

	logFile, err := os.OpenFile("/opt/nightlight/ipxe.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("iPXE server: cannot open log file: %v; logging to stderr", err)
		logFile = os.Stderr
	}
	ipxeLog := log.New(logFile, "", log.LstdFlags)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: ipxeLog, NoColor: true}))

	// Static boot files — kernels, initrds, iPXE binaries, etc.
	r.Handle("/files/*", http.StripPrefix("/files", http.FileServer(http.Dir(ipxeFilesDir))))

	// iPXE script by instance ID.
	r.Get("/ipxe/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var inst utils.Instance
		if err := db.One("ID", id, &inst); err != nil {
			ipxeLog.Printf("ipxe script not found: id=%s remote=%s", id, r.RemoteAddr)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		ipxeLog.Printf("serving iPXE script: id=%s instance=%s remote=%s", id, inst.Name, r.RemoteAddr)
		serveIPXE(w, &inst)
	})

	// iPXE script by NIC MAC address (colons or plain 12-hex).
	r.Get("/ipxe/mac/{mac}", func(w http.ResponseWriter, r *http.Request) {
		mac := chi.URLParam(r, "mac")
		if len(mac) == 12 && !strings.Contains(mac, ":") {
			mac = strings.Join([]string{mac[0:2], mac[2:4], mac[4:6], mac[6:8], mac[8:10], mac[10:12]}, ":")
		}
		inst := instanceByNICMac(db, mac)
		if inst == nil {
			ipxeLog.Printf("ipxe script not found: mac=%s remote=%s", mac, r.RemoteAddr)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		ipxeLog.Printf("serving iPXE script: mac=%s instance=%s remote=%s", mac, inst.Name, r.RemoteAddr)
		serveIPXE(w, inst)
	})

	// Kickstart by instance ID.
	r.Get("/kickstart/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var inst utils.Instance
		if err := db.One("ID", id, &inst); err != nil {
			ipxeLog.Printf("kickstart not found: id=%s remote=%s", id, r.RemoteAddr)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		ipxeLog.Printf("serving kickstart: id=%s instance=%s remote=%s", id, inst.Name, r.RemoteAddr)
		serveKickstart(w, &inst)
	})

	// Kickstart by NIC MAC address (colons or plain 12-hex).
	r.Get("/kickstart/mac/{mac}", func(w http.ResponseWriter, r *http.Request) {
		mac := chi.URLParam(r, "mac")
		if len(mac) == 12 && !strings.Contains(mac, ":") {
			mac = strings.Join([]string{mac[0:2], mac[2:4], mac[4:6], mac[6:8], mac[8:10], mac[10:12]}, ":")
		}
		inst := instanceByNICMac(db, mac)
		if inst == nil {
			ipxeLog.Printf("kickstart not found: mac=%s remote=%s", mac, r.RemoteAddr)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		ipxeLog.Printf("serving kickstart: mac=%s instance=%s remote=%s", mac, inst.Name, r.RemoteAddr)
		serveKickstart(w, inst)
	})

	srv := &http.Server{
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	ln, err := net.Listen("tcp", ipxeListenAddr)
	if err != nil {
		return err
	}

	// Unix socket for ipxeagent processes running in OVS-attached namespaces.
	// Stale socket from a previous run is removed first.
	os.Remove(ipxeSocketPath)
	if unixLn, unixErr := net.Listen("unix", ipxeSocketPath); unixErr != nil {
		ipxeLog.Printf("iPXE server: cannot listen on Unix socket %s: %v", ipxeSocketPath, unixErr)
	} else {
		os.Chmod(ipxeSocketPath, 0660)
		go func() {
			if err := srv.Serve(unixLn); err != nil && err != http.ErrServerClosed {
				ipxeLog.Printf("iPXE Unix socket error: %v", err)
			}
		}()
		ipxeLog.Printf("iPXE server also listening on unix:%s", ipxeSocketPath)
	}

	ipxeLog.Printf("iPXE server listening on %s (files: %s, log: /opt/nightlight/ipxe.log)", ipxeListenAddr, ipxeFilesDir)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			ipxeLog.Printf("iPXE server shutdown: %v", err)
		}
		os.Remove(ipxeSocketPath)
		return nil
	case err := <-errCh:
		return err
	}
}

// serveKickstart writes the instance's Kickstart content as plain text.
// If the field is empty the response is 404.
func serveKickstart(w http.ResponseWriter, inst *utils.Instance) {
	if inst.Kickstart == "" {
		http.Error(w, "no kickstart configured for this instance", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(inst.Kickstart))
}

// serveIPXE writes the instance's iPXE content as plain text.
// If the field is empty the response is 404.
func serveIPXE(w http.ResponseWriter, inst *utils.Instance) {
	if inst.IPXEScript == "" {
		http.Error(w, "no iPXE script configured for this instance", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(inst.IPXEScript))
}

// instanceByNICMac finds an instance whose PrimaryMacAddress or any NIC
// MacAddress matches mac (case-insensitive).
func instanceByNICMac(db *storm.DB, mac string) *utils.Instance {
	var instances []utils.Instance
	if err := db.All(&instances); err != nil {
		return nil
	}
	for i := range instances {
		if strings.EqualFold(instances[i].PrimaryMacAddress, mac) {
			return &instances[i]
		}
		for _, nic := range instances[i].Devices.NetworkInterfaces {
			if strings.EqualFold(nic.MacAddress, mac) {
				return &instances[i]
			}
		}
	}
	return nil
}
