package localcopy

import (
	"bufio"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const testFilename = "./testdata/test.txt"

type testdata string

func readTestdata(r io.Reader) (interface{}, error) {
	lines := bufio.NewScanner(r)
	var line string
	if lines.Scan() {
		line = lines.Text()
	}
	return line, nil
}

var serveTestdata = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, testFilename)
})

func TestLoadLocal(t *testing.T) {
	value, err := LoadLocal("./testdata/test.txt", readTestdata)
	if err != nil {
		t.Errorf("loading failed: %v", err)
		t.FailNow()
	}

	if len(value.(string)) != 51 {
		t.Errorf("expected a length of 51, but got %d", len(value.(string)))
	}
}

func TestLoadRemote(t *testing.T) {
	testServer := httptest.NewServer(serveTestdata)
	defer testServer.Close()

	value, err := LoadRemote(testServer.URL, readTestdata)
	if err != nil {
		t.Errorf("loading failed: %v", err)
		t.FailNow()
	}
	if len(value.(string)) != 51 {
		t.Errorf("expected a length of 51, but got %d: %q", len(value.(string)), value)
	}
}

func TestDownload(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "localcopy.TestDownload")
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
		t.FailNow()
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	testServer := httptest.NewServer(serveTestdata)
	defer testServer.Close()

	err = Download(testServer.URL, tempFile.Name(), readTestdata)
	if err != nil {
		t.Errorf("failed to download: %v", err)
		t.FailNow()
	}

	tempFileInfo, _ := os.Stat(tempFile.Name())
	testFileInfo, _ := os.Stat(testFilename)
	if tempFileInfo.Size() != testFileInfo.Size() {
		t.Errorf("expected file size %d, but got %d", testFileInfo.Size(), tempFileInfo.Size())
	}
}

func TestNeedsUpdate(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "localcopy.TestDownload")
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
		t.FailNow()
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())
	tempFileInfo, _ := os.Stat(tempFile.Name())

	timeToServe := time.Now()
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !timeToServe.IsZero() {
			w.Header().Add(httpLastModified, timeToServe.Format(httpTimeFormat))
		}
	}))
	defer testServer.Close()

	testCases := []struct {
		remoteFileTime time.Time
		needsUpdate    bool
	}{
		{time.Time{}, false},
		{tempFileInfo.ModTime().Add(-10 * time.Minute), false},
		{tempFileInfo.ModTime().Add(10 * time.Minute), true},
	}

	for _, testCase := range testCases {
		timeToServe = testCase.remoteFileTime

		needsUpdate, err := NeedsUpdate(testServer.URL, tempFile.Name())
		if timeToServe.IsZero() {
			if err == nil {
				t.Errorf("missing Last-Update header should raise an error")
			}
		} else {
			if err != nil {
				t.Errorf("failed to check for update: %v", err)
				t.FailNow()
			}
			if needsUpdate != testCase.needsUpdate {
				t.Errorf("expected needsUpdate %t, but got %t for %v", testCase.needsUpdate, needsUpdate, timeToServe.Sub(tempFileInfo.ModTime()))
			}
		}
	}
}
