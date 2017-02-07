package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kudinovdenis/csServer/logger"
	"github.com/kudinovdenis/csServer/searchAPI"
	"github.com/kudinovdenis/csServer/storage"
)

func receivePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Log(logger.LogLevelDefault, "URL not support method")
		return
	}
	error := r.ParseForm()
	if error != nil {
		logger.Logf(logger.LogLevelError, "%s", error.Error())
		return
	}
	error = r.ParseMultipartForm(1024)
	if error != nil {
		logger.Logf(logger.LogLevelError, "%s", error.Error())
		return
	}
	assetID := r.Form.Get("assetID")

	file, header, error := r.FormFile("photo")
	defer file.Close()
	if error != nil {
		logger.Logf(logger.LogLevelError, "%s", error.Error())
		return
	}
	logger.Logf(logger.LogLevelDefault, "Uploading file: %s", header.Filename)

	// Saving photo
	error = os.Mkdir("/tmp/images", os.ModePerm)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Directory is already exists %s", error.Error())
	}
	fileURL := "/tmp/images/" + assetID
	out, error := os.Create(fileURL)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant create file %s", error.Error())
		return
	}
	defer out.Close()
	_, error = io.Copy(out, file)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant copy file %s", error.Error())
		return
	}

	if storage.IsPhotoExists(assetID) {
		bytes, parseError := json.Marshal(storage.TagsForPhoto(assetID))
		if parseError != nil {
			logger.Logf(logger.LogLevelError, "Cant marshal JSON %s", error.Error())
			return
		}
		w.Write(bytes)
		return
	}

	// start searching
	info := searchAPI.InfoForPhoto(assetID)
	bytes, error := json.Marshal(info)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant marshal JSON %s", error.Error())
		return
	}
	logger.Logf(logger.LogLevelDefault, "info: %s", string(bytes))

	bytes, error = json.Marshal(info.Tags)
	if error != nil {
		logger.Logf(logger.LogLevelError, "Cant marshal JSON %s", error.Error())
		return
	}
	tags := []string{}
	for i := 0; i < len(info.Tags); i++ {
		tag := info.Tags[i].Name
		tags = append(tags, tag)
	}
	storage.SavePhoto(assetID, fileURL, tags)
	w.Write(bytes)
}

func main() {
	logger.Log(logger.LogLevelDefault, "Starting...")
	mysqlIP := flag.String("mysqlIP", "127.0.0.1", "pass an Mysql ip address")
	flag.Parse()
	logger.Logf(logger.LogLevelDefault, "Mysql server IP: %s", *mysqlIP)
	storage.InitDB("storage", *mysqlIP)
	// storage.FindTopTags(40)
	http.HandleFunc("/uploadImage", receivePost)
	error := http.ListenAndServe(":80", nil)
	if error != nil {
		fmt.Printf("%s", error.Error())
		return
	}
}
