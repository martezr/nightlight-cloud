package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/martezr/go-openvswitch/ovs"
	"github.com/martezr/nightlight-cloud/database"
	"github.com/martezr/nightlight-cloud/network"
	"github.com/ovn-org/libovsdb/client"

	"github.com/evangwt/go-vncproxy"
)

var dbhost = os.Getenv("DB_DIR")
var (
	db            *storm.DB
	networkClient client.Client
)

//go:embed webui/dist/*
var webui embed.FS

func main() {
	log.Println("nightlight-cloud 0.0.1")

	// Connect to the database
	db = database.StartDB(".")

	// Perform base configuration
	baseConfiguration()

	// Setup networking
	network.SetupBaseNetworking()

	configureDefaultNetworking()
	configureDefaultStorage()

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

	// Subnets
	r.Post("/api/v1/subnets", CreateSubnet)
	r.Get("/api/v1/subnets", ListSubnets)
	r.Delete("/api/v1/subnets/{id}", DeleteSubnet)

	// Instances
	r.Get("/api/v1/instances", ListInstances)
	r.Post("/api/v1/instances", CreateInstance)
	r.Delete("/api/v1/instances/{id}", DeleteInstance)
	r.Post("/api/v1/instances/{id}/restart", RestartInstance)
	r.Post("/api/v1/instances/{id}/sendkeys", SendInstanceConsoleKeys)

	// Datastores
	r.Get("/api/v1/datastores", ListDatastores)
	r.Post("/api/v1/datastores", CreateDatastore)
	r.Delete("/api/v1/datastores/{id}", DeleteDatastore)
	r.Get("/api/v1/datastores/{id}/files", ListDatastoreFiles)
	r.Post("/api/v1/datastores/{id}/fetch", DownloadDatastoreFile)

	r.Get("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"version":"0.0.1"}`))
	})

	//vncProxy := NewVNCProxy()
	//r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
	//	h := websocket.Handler(vncProxy.ServeWS)
	//	h.ServeHTTP(w, r)
	//})

	r.NotFound(NotFoundHandler)
	log.Println("Listening on port 80")

	if err := waitForPing("10.0.0.235", 60*time.Second); err != nil {
		log.Fatalf("Host 10.0.0.235 not reachable: %v", err)
	}
	log.Println("Host 10.0.0.235 is reachable, starting server")

	http.ListenAndServe("0.0.0.0:80", r)
}

// Set the system hostname
func setHostname(hostname string) error {
	err := os.WriteFile("/etc/hostname", []byte(hostname), 0644)
	if err != nil {
		return err
	}

	// Set the hostname immediately
	err = os.WriteFile("/proc/sys/kernel/hostname", []byte(hostname), 0644)
	if err != nil {
		return err
	}

	return nil
}

// set root password
func setRootPassword(password string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo 'root:%s' | chpasswd", password))
	return cmd.Run()
}

// configure root ssh access
func configureRootSSH() error {
	sshDir := "/root/.ssh"
	err := os.MkdirAll(sshDir, 0700)
	if err != nil {
		return err
	}

	// For demo purposes, we use a hardcoded public key. In production, consider generating a new key pair or using a secure method.
	publicKey := ""
	err = os.WriteFile(sshDir+"/authorized_keys", []byte(publicKey), 0600)
	if err != nil {
		return err
	}

	return nil
}

// enable ssh password authentication
func enableSSHPasswordAuth() error {
	// check if /etc/ssh/sshd_config.d/10-nightlight.conf exists, if not create it
	if _, err := os.Stat("/etc/ssh/sshd_config.d/10-nightlight.conf"); os.IsNotExist(err) {
		err = os.WriteFile("/etc/ssh/sshd_config.d/10-nightlight.conf", []byte("PasswordAuthentication yes\nPermitRootLogin yes\n"), 0644)
		if err != nil {
			return err
		}
		restartErr := restartSSHService()
		if restartErr != nil {
			return restartErr
		}
		return nil
	}
	return nil
}

// restart ssh service
func restartSSHService() error {
	cmd := exec.Command("sh", "-c", "rc-service sshd restart")
	return cmd.Run()
}

func baseConfiguration() {
	// make /opt/nightlight directory
	os.MkdirAll("/opt/nightlight/volumes", 0755)

	err := setHostname("nightlight-cloud")
	if err != nil {
		log.Fatalf("Error setting hostname: %v", err)
	}
	err = setRootPassword("nightlight")
	if err != nil {
		log.Fatalf("Error setting root password: %v", err)
	}
	err = configureRootSSH()
	if err != nil {
		log.Fatalf("Error configuring root SSH: %v", err)
	}

	err = enableSSHPasswordAuth()
	if err != nil {
		log.Fatalf("Error enabling SSH password authentication: %v", err)
	}
}

func NewVNCProxy() *vncproxy.Proxy {
	return vncproxy.New(&vncproxy.Config{
		LogLevel: vncproxy.DebugLevel,
		TokenHandler: func(r *http.Request) (addr string, err error) {
			// validate token and get forward vnc addr
			// ...
			addr = ":5901"
			return
		},
	})
}

func configureDefaultNetworking() {
	// Create a default VPC and subnet if they don't exist
	var vpcs []VPC
	err := db.All(&vpcs)
	if err != nil {
		log.Fatalf("Error fetching VPCs: %v", err)
	}
	if len(vpcs) == 0 {
		defaultVPC := VPC{
			ID:        "defaultvpc",
			Name:      "defaultvpc",
			CIDRBlock: "10.0.0.0/16",
		}
		db.Save(&defaultVPC)

		defaultSubnet := Subnet{
			ID:         "defaultsubnet",
			VPCId:      defaultVPC.ID,
			Name:       "defaultsubnet",
			CIDRBlock:  "10.0.0.0/24",
			BridgeName: "nightlight",
		}
		db.Save(&defaultSubnet)
	}

	// Create metadata network namespace and OVS interface
	ovsClient := ovs.New()
	ovsClient.VSwitch.AddPort("nightlight", "mddefaultvpc")
	ovsClient.VSwitch.Set.Interface("mddefaultvpc", ovs.InterfaceOptions{
		Type: "internal",
		ExternalIds: map[string]string{
			"iface-id":     "mddefaultvpc",
			"attached-mac": "32:6b:ce:89:41:42",
		},
	})

	err = network.CreateNetworkNamespace("mddefaultvpc", "32:6b:ce:89:41:42", "169.254.169.254")
	if err != nil {
		log.Fatalf("Error creating network namespace: %v", err)
	}

	/*
		// Create dhcp network namespace and OVS interface
		ovsClient.VSwitch.AddPort("nightlight", "dhdefaultvpc")
		ovsClient.VSwitch.Set.Interface("dhdefaultvpc", ovs.InterfaceOptions{
			Type: "internal",
			ExternalIds: map[string]string{
				"iface-id":     "dhdefaultvpc",
				"attached-mac": "32:6b:ce:89:41:43",
			},
		})

		err = network.CreateNetworkNamespace("dhdefaultvpc", "32:6b:ce:89:41:43", "169.254.169.253")
		if err != nil {
			log.Fatalf("Error creating network namespace: %v", err)
		}
	*/
}

func configureDefaultStorage() {
	// Create a default datastore if it doesn't exist
	var datastores []Datastore
	err := db.All(&datastores)
	if err != nil {
		log.Fatalf("Error fetching datastores: %v", err)
	}
	if len(datastores) == 0 {
		defaultDatastore := Datastore{
			ID:        "defaultdatastore",
			Name:      "defaultdatastore",
			LocalPath: "/opt/nightlight/volumes/defaultdatastore",
		}
		os.MkdirAll(defaultDatastore.LocalPath, 0755)
		db.Save(&defaultDatastore)
	}
}

// waitForPing pings addr until it responds or timeout elapses.
func waitForPing(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		// send a single ICMP echo request, wait 1s for reply
		cmd := exec.Command("ping", "-c", "1", "-W", "1", addr)
		if err := cmd.Run(); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for ping to %s", addr)
		}
		time.Sleep(1 * time.Second)
	}
}
