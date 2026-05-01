package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/hashicorp/go-hclog"
	"golang.org/x/net/websocket"

	"github.com/martezr/nightlight-cloud/database"
	"github.com/martezr/nightlight-cloud/metadatabackend"
	"github.com/martezr/nightlight-cloud/utils"

	"github.com/evangwt/go-vncproxy"
)

var (
	db *storm.DB
)

//go:embed webui/dist/*
var webui embed.FS

func main() {
	fmt.Println("==> Nightlight cloud:")
	fmt.Println("")
	mess := fmt.Sprintf(
		"%24s: %s",
		"API Address",
		fmt.Sprintf("0.0.0.0:%s", "80"))
	fmt.Println(mess)
	vauthVersion := fmt.Sprintf(
		"%24s: %s",
		"Version",
		"Nightlight cloud v0.0.1")
	fmt.Println(vauthVersion)
	fmt.Println("")
	fmt.Println("==> Nightlight cloud started! Log data will stream in below:")
	fmt.Println("")

	initialSetup()

	// Connect to the database
	db = database.StartDB("/opt/nightlight/nightlight.db")

	go metadatabackend.StartMetadataServer(context.Background(), db)
	go StartDHCPServer(context.Background(), db)

	defaultDatastore()
	// Setup HTTP server with routes
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Hosts
	r.Get("/api/v1/hosts", ListHosts)

	// VPCs
	r.Post("/api/v1/vpcs", CreateVPC)
	r.Get("/api/v1/vpcs/{id}", GetVPC)
	r.Get("/api/v1/vpcs", ListVpcs)
	r.Put("/api/v1/vpcs/{id}", UpdateVPC)
	r.Delete("/api/v1/vpcs/{id}", DeleteVPC)
	r.Get("/api/v1/vpcs/{id}/ips", ListAvailableIPs)
	r.Get("/api/v1/vpcs/{id}/graph", GetVPCGraph)
	r.Get("/api/v1/vpcs/{id}/flowlogs", GetVPCFlowLogs)
	r.Post("/api/v1/vpcs/{id}/ips/release", ReleaseVPCIPAddress)

	// Subnets
	r.Post("/api/v1/subnets", CreateSubnet)
	r.Get("/api/v1/subnets", ListSubnets)
	r.Delete("/api/v1/subnets/{id}", DeleteSubnet)

	// Instances
	r.Get("/api/v1/instances", ListInstances)
	r.Post("/api/v1/instances", CreateInstance)
	r.Get("/api/v1/instances/{id}", GetInstance)
	r.Delete("/api/v1/instances/{id}", DeleteInstance)
	r.Post("/api/v1/instances/{id}/restart", RestartInstance)
	r.Post("/api/v1/instances/{id}/stop", StopInstance)
	r.Post("/api/v1/instances/{id}/start", StartInstance)
	r.Post("/api/v1/instances/{id}/shutdown", ShutdownInstance)
	r.Post("/api/v1/instances/{id}/sendkeys", SendInstanceConsoleKeys)

	// Datastores
	r.Get("/api/v1/datastores", ListDatastores)
	r.Post("/api/v1/datastores", CreateDatastore)
	r.Get("/api/v1/datastores/{id}", GetDatastore)
	r.Delete("/api/v1/datastores/{id}", DeleteDatastore)
	r.Get("/api/v1/datastores/{id}/files", ListDatastoreFiles)
	r.Post("/api/v1/datastores/{id}/files", UploadDatastoreFile)
	r.Delete("/api/v1/datastores/{id}/files", DeleteDatastoreFile)
	r.Post("/api/v1/datastores/{id}/fetch", DownloadDatastoreFile)
	r.Get("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"version":"0.0.1"}`))
	})

	vncProxy := NewVNCProxy()
	r.Get("/ws/{id}", VNCWebsocketHandler(vncProxy))

	r.Get("/ssh", sshHandler)
	r.NotFound(NotFoundHandler)
	hclog.Default().Named("core").Info("Web UI available at http://<host-ip>:80/")
	hclog.Default().Named("core").Info("Default login: root / nightlight")

	hclog.Default().Named("core").Info("Listening on 0.0.0.0:80")

	srv := &http.Server{
		Addr:    "0.0.0.0:80",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	hclog.Default().Named("core").Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	hclog.Default().Named("core").Info("Server exited")
}

func VNCWebsocketHandler(proxy *vncproxy.Proxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		websocket.Handler(proxy.ServeWS).ServeHTTP(w, r)
	}
}

func NewVNCProxy() *vncproxy.Proxy {
	return vncproxy.New(&vncproxy.Config{
		LogLevel: vncproxy.DebugLevel,
		TokenHandler: func(r *http.Request) (addr string, err error) {
			id := chi.URLParam(r, "id")
			var instance utils.Instance
			err = db.One("ID", id, &instance)
			if err != nil {
				return
			}
			addr = fmt.Sprintf(":%d", instance.VNCPort)
			return
		},
	})
}

func defaultDatastore() {
	// Create a default datastore if it doesn't exist
	var datastores []Datastore
	err := db.All(&datastores)
	if err != nil {
		log.Fatalf("Error fetching datastores: %v", err)
	}
	if len(datastores) == 0 {
		defaultDatastore := Datastore{
			ID:            "defaultdatastore",
			Name:          "defaultdatastore",
			Description:   "Default nightlight cloud datastore",
			DatastoreType: "local",
			Path:          "/opt/nightlight/volumes/defaultdatastore",
			LocalPath:     "/opt/nightlight/volumes/defaultdatastore",
		}
		os.MkdirAll(defaultDatastore.LocalPath, 0755)
		db.Save(&defaultDatastore)
	}
}

func formatAndMount(device, mountPoint string) error {
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return err
	}

	if err := exec.Command("mkfs.ext4", device).Run(); err != nil {
		return err
	}

	if err := exec.Command("mount", device, mountPoint).Run(); err != nil {
		return err
	}

	return nil
}

func mountDisk(device, mountPoint string) error {
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return err
	}

	if err := exec.Command("mount", device, mountPoint).Run(); err != nil {
		return err
	}

	return nil
}

func checkDiskFormatted(diskName string) bool {
	// Run blkid command to get filesystem info
	cmd := exec.Command("blkid")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing blkid command: %s", err)
	}

	// Convert output to string and check for the disk name and filesystem
	outputStr := string(output)
	if strings.Contains(outputStr, diskName) {
		// Check if the disk has a filesystem (FSTYPE)
		if strings.Contains(outputStr, "TYPE=") {
			return true
		}
	}
	return false
}

// check if a disk is mounted at a given mount point
func isDiskMounted(mountPoint string) bool {
	cmd := exec.Command("mountpoint", "-q", mountPoint)
	err := cmd.Run()
	return err == nil
}

func initialSetup() {
	diskStatus := checkDiskFormatted("/dev/sda")
	if diskStatus {
		hclog.Default().Named("core").Info("Disk /dev/sda is already formatted.")
		if isDiskMounted("/data") {
			hclog.Default().Named("core").Info("Disk /dev/sda is already mounted at /data.")
			return
		}
		hclog.Default().Named("core").Info("Disk /dev/sda is not mounted. Attempting to mount...")
		err := mountDisk("/dev/sda", "/data")
		if err != nil {
			hclog.Default().Named("core").Error(fmt.Sprintf("Error mounting disk: %v", err))
		} else {
			hclog.Default().Named("core").Info("Successfully mounted disk to /data")
		}
	} else {
		hclog.Default().Named("core").Info("Disk /dev/sda is not formatted. Formatting and mounting...")
		// Create necessary directories
		dirs := []string{"/opt/nightlight"}
		for _, dir := range dirs {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				hclog.Default().Named("core").Error(fmt.Sprintf("Error creating directory %s: %v", dir, err))
			} else {
				hclog.Default().Named("core").Info(fmt.Sprintf("Successfully created directory: %s", dir))
			}
		}
		// Format and mount the disk
		err := formatAndMount("/dev/sda", "/opt/nightlight")
		if err != nil {
			hclog.Default().Named("core").Error(fmt.Sprintf("Error formatting and mounting disk: %v", err))
		} else {
			hclog.Default().Named("core").Info("Successfully formatted and mounted disk to /opt/nightlight")
		}
		// write the current version to a file
		err = os.WriteFile("/opt/nightlight/version.json", []byte(`{"version":"0.0.1"}`), 0644)
		if err != nil {
			hclog.Default().Named("core").Error(fmt.Sprintf("Error writing version file: %v", err))
		} else {
			hclog.Default().Named("core").Info("Successfully wrote version file")
		}
	}
}
