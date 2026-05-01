package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/hashicorp/go-hclog"
	"github.com/martezr/nightlight-cloud/utils"
)

type Datastore struct {
	ID            string              `json:"id" storm:"id,index"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	DatastoreType string              `json:"type"`
	Path          string              `json:"path"`
	LocalPath     string              `json:"localPath"`
	Tags          []map[string]string `json:"tags"`
}

type FileDetails struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	ModifiedDate string `json:"modifiedDate"`
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
	if err := json.NewDecoder(r.Body).Decode(&datastore); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	datastore.ID = "datastore-" + utils.IDGenerator(10)
	if datastore.DatastoreType == "local" {
		baseDirectory := fmt.Sprintf("/opt/nightlight/volumes/%s", datastore.ID)
		datastore.LocalPath = baseDirectory
		err := os.MkdirAll(datastore.LocalPath, 0755)
		if err != nil {
			hclog.Default().Named("core").Error(err.Error())
		}
	}
	if datastore.DatastoreType == "nfs" {
		baseDirectory := fmt.Sprintf("/opt/nightlight/volumes/%s", datastore.ID)
		datastore.LocalPath = baseDirectory
		err := os.MkdirAll(baseDirectory, 0755)
		if err != nil {
			hclog.Default().Named("core").Error(err.Error())
		}
		out := exec.Command("mount", "-t", "nfs", "-o", "vers=4", datastore.Path, datastore.LocalPath)
		err = out.Run()
		if err != nil {
			hclog.Default().Named("core").Error(err.Error())
		}
	}
	err := db.Save(&datastore)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
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

func GetDatastore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var datastore Datastore
	err := db.One("ID", id, &datastore)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(datastore))
}

func DeleteDatastore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var datastore Datastore
	err := db.One("ID", id, &datastore)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	if datastore.DatastoreType == "local" {
		err := os.RemoveAll(datastore.LocalPath)
		if err != nil {
			hclog.Default().Named("core").Error(fmt.Sprintf("failed to delete local folder: %v", err))
		}
	}

	if datastore.DatastoreType == "nfs" {
		out := exec.Command("umount", datastore.LocalPath)
		err = out.Run()
		if err != nil {
			hclog.Default().Named("core").Error(fmt.Sprintf("failed to unmount nfs: %v", err))
		}
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
	err = filepath.Walk(datastore.LocalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			hclog.Default().Named("core").Error(err.Error())
			return err
		}
		info.Size()
		var file FileDetails
		file.Name = info.Name()
		file.Path = path
		file.Size = info.Size()
		file.ModifiedDate = info.ModTime().String()
		files = append(files, file)
		return nil
	})
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(files))
}

func UploadDatastoreFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var datastore Datastore
	if err := db.One("ID", id, &datastore); err != nil {
		http.Error(w, "datastore not found", http.StatusNotFound)
		return
	}

	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "expected multipart/form-data request", http.StatusBadRequest)
		return
	}

	part, err := mr.NextPart()
	if err != nil {
		http.Error(w, "no file in request", http.StatusBadRequest)
		return
	}
	defer part.Close()

	filename := filepath.Base(part.FileName())
	if filename == "" || filename == "." {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}

	destPath := filepath.Join(datastore.LocalPath, filename)
	f, err := os.Create(destPath)
	if err != nil {
		hclog.Default().Named("core").Error(fmt.Sprintf("failed to create file: %v", err))
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := io.Copy(f, part); err != nil {
		os.Remove(destPath)
		hclog.Default().Named("core").Error(fmt.Sprintf("upload failed: %v", err))
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"name": filename})
}

func DownloadDatastoreFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var datastore Datastore
	err := db.One("ID", id, &datastore)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	var downloadFile DownloadFile
	if err := json.NewDecoder(r.Body).Decode(&downloadFile); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(datastore.LocalPath, filepath.Base(downloadFile.Name))
	hclog.Default().Named("core").Info(fmt.Sprintf("Downloading file to: %s", filePath))
	utils.DownloadFile(downloadFile.URL, filePath)

	payload := `{"status":"success"}`
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(payload))
}

func DeleteDatastoreFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var datastore Datastore
	if err := db.One("ID", id, &datastore); err != nil {
		http.Error(w, "datastore not found", http.StatusNotFound)
		return
	}

	var payload struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Name == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(datastore.LocalPath, filepath.Base(payload.Name))
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		hclog.Default().Named("core").Error(fmt.Sprintf("failed to delete file: %v", err))
		http.Error(w, "failed to delete file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func FindDatastoreByID(id string) (datastore Datastore) {
	err := db.One("ID", id, &datastore)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	return datastore
}
