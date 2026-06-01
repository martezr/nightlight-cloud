package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	db = database.StartDB("/opt/nightlight")

	go metadatabackend.StartMetadataServer(context.Background(), db)
	go StartDHCPServer(context.Background(), db)
	go StartIPXEServer(context.Background(), db)

	defaultDatastore()
	initDefaultUser()
	// Add logic to check if the wan router instance already exists before creating it
	// check if /opt/nightlight/volumes/defaultdatastore/wanrouter already exists, if not create the wan router instance and store its disk there
	if _, err := os.Stat("/opt/nightlight/volumes/defaultdatastore/wanrouter"); os.IsNotExist(err) {
		createWanRouter()
	}

	// Setup HTTP server with routes
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Public auth routes
	r.Post("/api/v1/auth/login", LoginHandler)
	r.Post("/api/v1/auth/logout", LogoutHandler)
	r.Get("/api/v1/auth/me", CurrentUserHandler)
	r.Put("/api/v1/auth/password", ChangePasswordHandler)

	// Protected API routes
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		// Hosts
		r.Get("/api/v1/hosts", ListHosts)

		// Bridges (live OVS query, read-only)
		r.Get("/api/v1/bridges", ListBridges)

		// Sites
		r.Post("/api/v1/sites", CreateSite)
		r.Get("/api/v1/sites/{id}", GetSite)
		r.Get("/api/v1/sites", ListSites)
		r.Put("/api/v1/sites/{id}", UpdateSite)
		r.Delete("/api/v1/sites/{id}", DeleteSite)
		r.Get("/api/v1/sites/{id}/switches", ListSwitchesBySite)

		// Switches
		r.Post("/api/v1/switches", CreateSwitch)
		r.Get("/api/v1/switches", ListSwitches)
		r.Get("/api/v1/switches/{id}", GetSwitch)
		r.Put("/api/v1/switches/{id}", UpdateSwitch)
		r.Delete("/api/v1/switches/{id}", DeleteSwitch)

		// Subnets
		r.Post("/api/v1/subnets", CreateSubnet)
		r.Get("/api/v1/subnets", ListSubnets)
		r.Delete("/api/v1/subnets/{id}", DeleteSubnet)
		r.Get("/api/v1/subnets/{id}", GetSubnet)
		r.Put("/api/v1/subnets/{id}", UpdateSubnet)
		r.Get("/api/v1/subnets/{id}/ips", ListAvailableIPs)
		r.Post("/api/v1/subnets/{id}/ips/release", ReleaseSubnetIPAddress)

		// Instances
		r.Get("/api/v1/instances", ListInstances)
		r.Post("/api/v1/instances", CreateInstance)
		r.Get("/api/v1/instances/{id}", GetInstance)
		r.Delete("/api/v1/instances/{id}", DeleteInstance)
		r.Post("/api/v1/instances/{id}/restart", RestartInstance)
		r.Post("/api/v1/instances/{id}/reset", ResetInstance)
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

		// Version
		r.Get("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"version":"0.0.1"}`))
		})
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
