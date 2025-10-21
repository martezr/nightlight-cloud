package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/hashicorp/go-hclog"
	"github.com/martezr/nightlight-cloud/utils"
)

type Image struct {
	ID              string                   `json:"id" storm:"id,index"`
	Version         int64                    `json:"version"`
	Description     string                   `json:"description"`
	Location        string                   `json:"location"`
	OperatingSystem string                   `json:"operatingSystem"`
	Tags            []map[string]interface{} `json:"tags"`
}

func CreateImage(w http.ResponseWriter, r *http.Request) {
	var image Image
	_ = json.NewDecoder(r.Body).Decode(&image)
	image.ID = "image-" + utils.IDGenerator(10)
	db.Save(&image)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(image))
}

func ListImages(w http.ResponseWriter, r *http.Request) {
	var images []Image
	err := db.All(&images)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(images))
}

func DeleteImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var image Image
	err := db.One("ID", id, &image)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	err = db.DeleteStruct(&image)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

func GetImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var image Image
	err := db.One("ID", id, &image)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(image))
}

func FindImageByID(id string) (image Image) {
	err := db.One("ID", id, &image)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	return image
}
