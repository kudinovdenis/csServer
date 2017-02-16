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
	"github.com/nu7hatch/gouuid"
	"crypto/md5"
	"io/ioutil"
	//"encoding/hex"
	"encoding/hex"
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

	file, header, err := r.FormFile("photo")
	defer file.Close()
	if err != nil {
		logger.Logf(logger.LogLevelError, "%s", err.Error())
		return
	}
	logger.Logf(logger.LogLevelDefault, "Uploading file: %s", header.Filename)
	filename := generateRandomFilename()

	// Saving photo
	err = os.MkdirAll("~/tmp/images", os.ModePerm)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant create directory. %s", err.Error())
	}
	fileURL := "~/tmp/images/" + filename
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

	fileMD5 := md5FromFile(fileURL)

	if newStorage.IsImageExists(fileMD5) {
		tags := newStorage.FindTagsForImage(fileMD5)
		bytes, parseError := json.Marshal(tags)
		if parseError != nil {
			logger.Logf(logger.LogLevelError, "Cant marshal JSON %s", err.Error())
			return
		}
		w.Write(bytes)
		return
	}

	// start searching
	info := searchAPI.InfoForPhoto(fileURL)
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
	newStorage.SaveImage(fileMD5, fileURL, tags)
	foundTags := newStorage.FindTagsForImage(fileMD5)
	bytes, parseError := json.Marshal(foundTags)
	if parseError != nil {
		logger.Logf(logger.LogLevelError, "Cant marshal JSON %s", err.Error())
		return
	}
	w.Write(bytes)
}

func generateRandomFilename() string {
	id, err := uuid.NewV4()
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant generate UUID. %s", err.Error())
	}
	return id.String()
}

func md5FromFile(filePath string) string {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant read file %s. %s", filePath, err)
	}
	hasher := md5.New()
	hasher.Write(b)
	return hex.EncodeToString(hasher.Sum(nil))
}

func main() {
	logger.Log(logger.LogLevelDefault, "Starting...")
	newStorage.InitDB("storage")
	// storage.InitDB("storage")
	// storage.FindTopTags(40)
	http.HandleFunc("/uploadImage", receivePost)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
}
