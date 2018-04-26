/*
Package localcopy allows to manage a local copy of a resource that is available through HTTP(s).

It provides functions do download the resource and to check if an update of the local copy
is necessary. The update check is done using a HEAD request and comparing the last
modification date of the local copy with the Last-Modified header of the HTTP response.
*/
package localcopy

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const httpTimeFormat = time.RFC1123
const httpLastModified = "Last-Modified"

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

// ReadFunc reads a value using the given reader.
type ReadFunc func(r io.Reader) (interface{}, error)

// LoadLocal loads the local copy from the local file system.
func LoadLocal(localFilename string, read ReadFunc) (interface{}, error) {
	file, err := os.Open(localFilename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	in := bufio.NewReader(file)
	value, err := read(in)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// LoadRemote loads the resource from the given remote location.
func LoadRemote(remoteURL string, read ReadFunc) (interface{}, error) {
	resp, err := httpClient.Get(remoteURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	in := bufio.NewReader(resp.Body)
	value, err := read(in)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Download downloads the resource from the given remote URL and stores it locally.
func Download(remoteURL, localFilename string, read ReadFunc) error {
	response, err := httpClient.Get(remoteURL)
	if err != nil {
		return fmt.Errorf("failed to download resource: %v", err)
	}
	defer response.Body.Close()

	os.MkdirAll(filepath.Dir(localFilename), os.ModePerm)
	localFile, err := os.Create(localFilename)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, response.Body)
	if err != nil {
		return fmt.Errorf("failed to store resource locally: %v", err)
	}

	return nil
}

// NeedsUpdate checks whether the local copy needs to be updated from the given remote URL.
func NeedsUpdate(remoteURL, localFilename string) (bool, error) {
	response, err := httpClient.Head(remoteURL)
	if err != nil {
		return false, err
	}
	var lastModified time.Time
	if lastModifiedHeader, ok := response.Header[httpLastModified]; ok {
		if len(lastModifiedHeader) == 0 {
			return false, fmt.Errorf("Last-Modified header is empty")
		}

		lastModified, err = time.Parse(httpTimeFormat, lastModifiedHeader[0])
		if err != nil {
			return false, fmt.Errorf("cannot parse Last-Modified header: %v", err)
		}
	} else {
		return false, fmt.Errorf("response does not contain a Last-Modified header")
	}

	localFileInfo, err := os.Stat(localFilename)
	if os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	return lastModified.After(localFileInfo.ModTime()), nil
}

// Update updates the local copy from the given remote URL, but only if an update is needed.
func Update(remoteURL, localFilename string, read ReadFunc) (bool, error) {
	needsUpdate, err := NeedsUpdate(remoteURL, localFilename)
	if err != nil {
		return false, err
	}

	if !needsUpdate {
		return false, nil
	}
	return true, Download(remoteURL, localFilename, read)
}
