package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kudinovdenis/csServer/logger"
	"github.com/kudinovdenis/csServer/newStorage"
	"github.com/kudinovdenis/csServer/searchAPI"
)

func receivePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Log(logger.LogLevelDefault, "URL not support method")
		return
	}
	err := r.ParseForm()
	if err != nil {
		logger.Logf(logger.LogLevelError, "%s", err.Error())
		return
	}
	err = r.ParseMultipartForm(1024)
	if err != nil {
		logger.Logf(logger.LogLevelError, "%s", err.Error())
		return
	}
	assetID := r.Form.Get("assetID")

	file, header, err := r.FormFile("photo")
	defer file.Close()
	if err != nil {
		logger.Logf(logger.LogLevelError, "%s", err.Error())
		return
	}
	logger.Logf(logger.LogLevelDefault, "Uploading file: %s", header.Filename)

	// Saving photo
	err = os.MkdirAll("~/tmp/images", os.ModePerm)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Directory is already exists %s", err.Error())
	}
	fileURL := "~/tmp/images/" + assetID
	out, err := os.Create(fileURL)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant create file %s", err.Error())
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant copy file %s", err.Error())
		return
	}

	if newStorage.IsImageExists(assetID) {
		tags := newStorage.FindTagsForImage(assetID)
		bytes, parseError := json.Marshal(tags)
		if parseError != nil {
			logger.Logf(logger.LogLevelError, "Cant marshal JSON %s", err.Error())
			return
		}
		w.Write(bytes)
		return
	}

	// start searching
	info := searchAPI.InfoForPhoto(assetID)
	bytes, err := json.Marshal(info)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant marshal JSON %s", err.Error())
		return
	}
	logger.Logf(logger.LogLevelDefault, "info: %s", string(bytes))

	bytes, err = json.Marshal(info.Tags)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant marshal JSON %s", err.Error())
		return
	}
	tags := newStorage.SaveTags(info.Tags)
	newStorage.SaveImage(assetID, fileURL, tags)
	w.Write(bytes)
}

func processSearchRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	searchAPI.SendQueryToLUIS(query)
}

func main() {
	logger.Log(logger.LogLevelDefault, "Starting...")
	newStorage.InitDB("storage")
	// storage.InitDB("storage")
	// storage.FindTopTags(40)
	http.HandleFunc("/uploadImage", receivePost)
	http.HandleFunc("/search", processSearchRequest)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
}
