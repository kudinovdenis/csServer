package searchAPI

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kudinovdenis/csServer/logger"
)

// SearchResponse ... search response
type SearchResponse struct {
	Tags      []Tag  `json:"tags"`
	RequestID string `json:"requestId"`
	Metadata  struct {
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Format string `json:"format"`
	} `json:"metadata"`
}

// Tag ... tag
type Tag struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	Hint       string  `json:"hint,omitempty"`
}

// InfoForPhoto ... Function assume that photo is already stored in PATH:
func InfoForPhoto(assetID string) SearchResponse {
	logger.Log(logger.LogLevelDefault, "Start microsoft api request")
	client := http.Client{}
	localURL := "~/tmp/images/" + assetID
	//
	file, err := os.Open(localURL)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant open file %s. %s.", localURL, err.Error())
		return SearchResponse{}
	}
	defer file.Close()
	//
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("", filepath.Base(localURL))
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant create reader %s. %s.", localURL, err.Error())
		return SearchResponse{}
	}
	_, err = io.Copy(part, file)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant copy part %s. %s.", localURL, err.Error())
		return SearchResponse{}
	}
	err = writer.Close()
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant close writer %s.", err.Error())
		return SearchResponse{}
	}

	request, err := http.NewRequest("POST", "https://westus.api.cognitive.microsoft.com/vision/v1.0/analyze?visualFeatures=Tags", body)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant create request. %s.", err.Error())
		return SearchResponse{}
	}

	request.Header.Add("Ocp-Apim-Subscription-Key", "c62e22612f8f4b47974faa0a906789f8")
	request.Header.Add("Content-Type", writer.FormDataContentType())
	logger.LogRequest(request, false)
	response, err := client.Do(request)

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant parse response %s.", err.Error())
	}

	logger.LogResponse(*response, responseBody)
	info := parseResponse(responseBody)
	return info
}

func parseResponse(response []byte) SearchResponse {
	var parsedResponse SearchResponse
	err := json.Unmarshal(response, &parsedResponse)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant parse json %s.", err.Error())
	}
	return parsedResponse
}
