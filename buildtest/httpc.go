/*
 *  Copyright (C) 2011-2021 Red Hat, Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *          http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package buildtest

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

// ContentType for RFC http content type (parts)
const (
	ContentTypePlain = "text/plain"
	ContentTypeHTML  = "text/html"

	ContentTypeJSON = "application/json"
	ContentTypeXML  = "application/xml"

	ContentTypeZip    = "application/zip"
	ContentTypeStream = "application/octet-stream"
	CottentTypeJar    = "application/java-archive"
)

// Status code for RFC http response status code (parts)
const (
	StatusOK        = http.StatusOK
	StatusCreated   = http.StatusCreated
	StatusAccepted  = http.StatusAccepted
	StatusNoContent = http.StatusNoContent

	StatusMultipleChoices  = http.StatusMultipleChoices
	StatusMovedPermanently = http.StatusMovedPermanently
	StatusFound            = http.StatusFound
	StatusSeeOther         = http.StatusSeeOther
	StatusNotModified      = http.StatusNotModified
	StatusUseProxy         = http.StatusUseProxy

	StatusBadRequest        = http.StatusBadRequest
	StatusUnauthorized      = http.StatusUnauthorized
	StatusForbidden         = http.StatusForbidden
	StatusNotFound          = http.StatusNotFound
	StatusMethodNotAllowed  = http.StatusMethodNotAllowed
	StatusNotAcceptable     = http.StatusNotAcceptable
	StatusProxyAuthRequired = http.StatusProxyAuthRequired
	StatusRequestTimeout    = http.StatusRequestTimeout
	StatusConflict          = http.StatusConflict

	StatusInternalServerError = http.StatusInternalServerError
	StatusNotImplemented      = http.StatusNotImplemented
	StatusBadGateway          = http.StatusBadGateway
	StatusServiceUnavailable  = http.StatusServiceUnavailable
	StatusGatewayTimeout      = http.StatusGatewayTimeout

	StatusUnknown = -1
)

// Methods for RFC http methods
const (
	MethodGet     = http.MethodGet
	MethodHead    = http.MethodHead
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodPatch   = http.MethodPatch
	MethodDelete  = http.MethodDelete
	MethodOptions = http.MethodOptions
)

const NotStoreFile = ""

type errorHandler func()

type Authenticate func(request *http.Request) error

//GetHost gets the hostname from a url string
func GetHost(URLString string) string {
	u, err := url.Parse(URLString)
	if err != nil {
		return ""
	}

	return u.Hostname()
}

//GetPort gets the port from a url string
func GetPort(URLString string) string {
	u, err := url.Parse(URLString)
	if err != nil {
		return "-1"
	}

	return u.Port()
}

func GetRespAsPlaintext(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", newHTTPError(err.Error(), 0)
	}
	defer resp.Body.Close()

	status, statusCode := resp.Status, resp.StatusCode

	if statusCode == StatusUnauthorized {
		fmt.Print("This API needs authorization, seems you need to get accesss token first. Please have a look at login command.\n\n")
		return "", newHTTPError(status, statusCode)
	}

	if statusCode > StatusBadRequest {
		return "", newHTTPError(status, statusCode)
	}

	if !strings.Contains(resp.Header.Get("content-type"), ContentTypePlain) {
		Printlnf("Warning: response from %s may be not a plain text content, may cause problem", url)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", newHTTPError(err.Error(), 0)
	}

	return string(b), nil
}

// HTTPRequest do raw http request with method, input data and headers. If url is trying to access bin content(like file), can use filename parameter to specify where to store this file as.
// Parameters: request url; request method; authentication method; if need response content; data payload to send(POST or PUT); headers to send; the file location to store if response is a binary download; if print verbose log message for debugging
// Returns: content as string, response status code as int, if succeeded as bool
func HTTPRequest(url, method string, auth Authenticate, needResult bool, dataPayload io.Reader, headers map[string]string, filename string, verbose bool) (string, int, bool) {
	client := &http.Client{}
	respText := ""
	req, err := http.NewRequest(method, url, dataPayload)
	if err != nil {
		fmt.Println(err)
		return respText, StatusUnknown, false
	}
	if headers != nil && len(headers) > 0 {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	if auth != nil {
		err := auth(req)
		if err != nil {
			fmt.Println(err)
			return "", StatusUnknown, false
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return respText, StatusUnknown, false
	}

	if resp.StatusCode >= 400 {
		Printlnf("%s request not success for %s, status: %s, return code: %v", method, url, resp.Status, resp.StatusCode)
		return respText, resp.StatusCode, false
	}

	if needResult {
		content, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			panic(err)
		}

		resp.Body.Close()

		return string(content), resp.StatusCode, true
	}

	return respText, resp.StatusCode, true
}

func newHTTPError(message string, statusCode int) HTTPError {
	return HTTPError{message, statusCode}
}

//HTTPError represents a generic http problem
type HTTPError struct {
	Message    string
	StatusCode int
}

func (err HTTPError) Error() string {
	return err.Message
}

func isBinContent(headers http.Header) bool {
	contentType := headers.Get("Content-Type")
	// fmt.Println(contentType)
	if strings.HasPrefix(contentType, "text") {
		return false
	}
	if contentType == ContentTypeJSON || contentType == ContentTypeXML {
		return false
	}

	return true
}

func downloadFile(url, storeFileName string, verbose bool) {
	fmt.Printf("Downloading %s\n", url)
	client := &http.Client{}
	req, err := http.NewRequest(MethodGet, url, nil)
	if err != nil {
		fmt.Printf("Can not download file %s, err: %s", url, err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Can not download file %s, err: %s", url, err)
		return
	}

	PrintVerbose("The api is trying to download a file\n", verbose)
	conDispo := resp.Header.Get("Content-Disposition")
	filePath := ""
	if !IsEmptyString(storeFileName) {
		filePath = strings.TrimSpace(storeFileName)
	} else {
		if !IsEmptyString(conDispo) {
			start := strings.Index(conDispo, "filename")
			filePath = conDispo[start:]
			splitted := strings.Split(filePath, "=")
			filePath = splitted[1]
		} else {
			filePath = path.Base(url)
		}
		filePath = "./" + filePath
	}

	// Check and create the file
	for FileOrDirExists(filePath) {
		filePath = filePath + ".1"
	}
	out, err := os.Create(filePath)
	defer out.Close()
	if err != nil {
		fmt.Println("Warning: cannot download file due to io error!")
	} else {
		PrintVerbose("Download started.\n", verbose)

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			fmt.Println("Warning: cannot download file due to io error!")
		} else {
			fmt.Println()
			PrintVerbose(fmt.Sprintf("\nFile downloaded as %s\n", filePath), verbose)
		}
	}

	fmt.Printf("Downloaded %s\n", url)
}
