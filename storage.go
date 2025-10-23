package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/hashicorp/go-hclog"
	"github.com/martezr/nightlight-cloud/utils"
)

type Datastore struct {
	ID            string                   `json:"id" storm:"id,index"`
	Name          string                   `json:"name"`
	DatastoreType string                   `json:"type"`
	Path          string                   `json:"path"`
	LocalPath     string                   `json:"localPath"`
	Tags          []map[string]interface{} `json:"tags"`
}

type FileDetails struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type DownloadFile struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Path string `json:"path"`
}

type DatastoreFilesListResponse struct {
	Files []FileDetails `json:"files"`
}

type DatastoreGetResponse struct {
	Datastore Datastore `json:"datastore"`
}

func CreateDatastore(w http.ResponseWriter, r *http.Request) {
	var datastore Datastore
	_ = json.NewDecoder(r.Body).Decode(&datastore)
	datastore.ID = "datastore-" + utils.IDGenerator(10)
	if datastore.DatastoreType == "local" {
		baseDirectory := fmt.Sprintf("/opt/nightlight/volumes/%s", datastore.ID)
		datastore.LocalPath = baseDirectory
		err := os.MkdirAll(datastore.LocalPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	if datastore.DatastoreType == "nfs" {
		baseDirectory := fmt.Sprintf("/opt/nightlight/volumes/%s", datastore.ID)
		datastore.LocalPath = baseDirectory
		err := os.MkdirAll(baseDirectory, 0755)
		if err != nil {
			log.Fatal(err)
		}
		out := exec.Command("mount", "-t", "nfs", "-o", "vers=4", datastore.Path, datastore.LocalPath)
		err = out.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
	db.Save(&datastore)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(datastore))
}

func ListDatastores(w http.ResponseWriter, r *http.Request) {
	var datastores []Datastore
	err := db.All(&datastores)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(datastores))
}

func DeleteDatastore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var datastore Datastore
	err := db.One("ID", id, &datastore)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	err = db.DeleteStruct(&datastore)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

func ListDatastoreFiles(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var datastore Datastore
	err := db.One("ID", id, &datastore)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	var files []FileDetails
	err = filepath.Walk(datastore.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		info.Size()
		var file FileDetails
		file.Name = info.Name()
		file.Path = path
		file.Size = info.Size()
		files = append(files, file)
		fmt.Printf("dir: %v: name: %s\n", info.IsDir(), path)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	payload := DatastoreFilesListResponse{Files: files}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(payload))
}

func DownloadDatastoreFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var datastore Datastore
	err := db.One("ID", id, &datastore)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	var downloadFile DownloadFile
	_ = json.NewDecoder(r.Body).Decode(&downloadFile)

	filePath := fmt.Sprintf("%s/%s", datastore.Path, downloadFile.Name)
	fmt.Println(datastore.Path)
	fmt.Println(downloadFile)
	fmt.Println(filePath)
	utils.DownloadFile(downloadFile.URL, filePath)

	payload := `{"status":"success"}`
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(payload))
}

func DeleteDatastoreFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var datastore Datastore
	err := db.One("ID", id, &datastore)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

func FindDatastoreByID(id string) (datastore Datastore) {
	err := db.One("ID", id, &datastore)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	return datastore
}
