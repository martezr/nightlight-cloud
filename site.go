package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/hashicorp/go-hclog"
	"github.com/martezr/go-openvswitch/ovs"
	"github.com/martezr/nightlight-cloud/utils"
)

// Site represents a physical site
type Site struct {
	ID           string   `json:"id" storm:"id,index"`
	Name         string   `json:"name" storm:"index"`
	Location     string   `json:"location"`
	Type         string   `json:"type"`
	Topology     string   `json:"topology"`
	Bridges      []string `json:"bridges"`
	WanBandwidth int      `json:"wanBandwidth"`
	Description  string   `json:"description"`
}

// ListVirtualNetworks lists all virtual networks
func ListSites(w http.ResponseWriter, r *http.Request) {
	var sites []Site
	err := db.All(&sites)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(sites))
}

// CreateSite creates a new site
func CreateSite(w http.ResponseWriter, r *http.Request) {
	var site Site
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	site.ID = "site-" + utils.IDGenerator(3)

	// Always provision the core switch for the site.
	if _, err := provisionSwitch(site.Name+" Core", "", site.ID, site.ID, "core", nil); err != nil {
		hclog.Default().Named("core").Error(err.Error())
		http.Error(w, "failed to create core switch", http.StatusInternalServerError)
		return
	}

	createTier1Router(site.Name, site.ID)

	ovsClient := ovs.New()
	switch site.Topology {
	case "spine-leaf":
		site.Bridges = []string{site.ID + "-lf1", site.ID + "-lf2"}
		for _, bridge := range site.Bridges {
			if _, err := provisionSwitch(bridge, "", bridge, site.ID, "leaf", nil); err != nil {
				hclog.Default().Named("core").Error(err.Error())
			}
			ovsClient.VSwitch.AddPort(site.ID, fmt.Sprintf("%s-dl-patch", bridge))
			ovsClient.VSwitch.AddPort(bridge, fmt.Sprintf("%s-ul-patch", bridge))
			ovsClient.VSwitch.Set.Interface(fmt.Sprintf("%s-dl-patch", bridge), ovs.InterfaceOptions{
				Type: "patch",
				Peer: fmt.Sprintf("%s-ul-patch", bridge),
			})
			ovsClient.VSwitch.Set.Interface(fmt.Sprintf("%s-ul-patch", bridge), ovs.InterfaceOptions{
				Type: "patch",
				Peer: fmt.Sprintf("%s-dl-patch", bridge),
			})
		}
	case "robo":
		site.Bridges = []string{site.ID + "-rb1", site.ID + "-rb2"}
		for _, bridge := range site.Bridges {
			if _, err := provisionSwitch(bridge, "", bridge, site.ID, "access", nil); err != nil {
				hclog.Default().Named("core").Error(err.Error())
			}
			ovsClient.VSwitch.AddPort(bridge, site.ID)
		}
	default:
		site.Topology = "single-bridge"
		site.Bridges = []string{site.ID}
	}

	db.Save(&site)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(site))
}

// GetSite fetches a site by ID and returns it
func GetSite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var site Site
	err := db.One("ID", id, &site)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(site))
}

// UpdateSite updates an existing site by ID
func UpdateSite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var site Site
	err := db.One("ID", id, &site)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	var data Site
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	data.ID = site.ID
	err = db.Update(&data)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

// DeleteSite deletes an existing site by ID
func DeleteSite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var site Site
	err := db.One("ID", id, &site)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	ovsClient := ovs.New()
	ovsClient.VSwitch.DeleteBridge(site.ID)
	err = db.DeleteStruct(&site)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

// FindSiteByName finds a site by name and returns it
func FindSiteByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var site Site
	err := db.One("Name", name, &site)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(site))
}

func FindSiteByID(id string) (site Site) {
	err := db.One("ID", id, &site)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	return site
}
