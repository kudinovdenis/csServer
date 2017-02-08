package logger

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Error logging level
const (
	LogLevelError   = iota
	LogLevelDefault = iota
)

// Log ... Append custom log identifier: [E!], [ ]
func Log(level int, s string) {
	if level == LogLevelError {
		fmt.Println("[E!] " + s)
	} else if level == LogLevelDefault {
		fmt.Println("[ ] " + s)
	}
}

// Logf ... Custom logging with format. Append custom log identifier: [E!], [ ]
func Logf(level int, format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	Log(level, message)
}

// LogResponse ... log response
func LogResponse(res http.Response, body []byte) {
	bodyString := string(body)
	message := fmt.Sprintf("Response for [%s]: Status: %s. Body: %s", res.Request.URL.String(), strconv.Itoa(res.StatusCode), bodyString)
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		Log(LogLevelDefault, message)
	} else {
		Log(LogLevelError, message)
	}
}

// LogRequest ... log request
func LogRequest(req *http.Request, logBody bool) {
	headersString := ""
	for k, v := range req.Header {
		headersString = headersString + fmt.Sprintf("[key:%s value:%s] ", k, v)
	}
	message := fmt.Sprintf("Request started: %s [%s]: Headers: %s.", req.Method, req.URL.String(), headersString)
	if logBody {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			Logf(LogLevelError, "Cant parse response %s.", err.Error())
		}
		bodyString := string(body[:])
		message += fmt.Sprintf(" Body: %s", bodyString)
	}
	Log(LogLevelDefault, message)
}
