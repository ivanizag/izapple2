package izapple2

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"embed"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ivanizag/izapple2/storage"
)

const (
	internalPrefix = "<internal>/"
	embedPrefix    = "resources/"
	httpPrefix     = "http://"
	httpsPrefix    = "https://"
)

//go:embed resources
var internalFiles embed.FS

func isInternalResource(filename string) bool {
	return strings.HasPrefix(filename, internalPrefix)
}

func isHTTPResource(filename string) bool {
	return strings.HasPrefix(filename, httpPrefix) ||
		strings.HasPrefix(filename, httpsPrefix)
}

// LoadResource loads in memory a file from the filesystem, http or embedded
func LoadResource(filename string) ([]uint8, bool, error) {
	// Remove quotes if surrounded by them
	if strings.HasPrefix(filename, "\"") && strings.HasSuffix(filename, "\"") {
		filename = filename[1 : len(filename)-1]
	}

	// Expand the tilde if prefixed by it
	if strings.HasPrefix(filename, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			filename = home + filename[1:]
		}
	}

	var writeable bool
	var file io.Reader
	if isInternalResource(filename) {
		// load from embedded resource
		resource := embedPrefix + strings.TrimPrefix(filename, internalPrefix)
		resourceFile, err := internalFiles.Open(resource)
		if err != nil {
			return nil, false, err
		}
		defer resourceFile.Close()
		file = resourceFile
		writeable = false

	} else if isHTTPResource(filename) {
		response, err := http.Get(filename)
		if err != nil {
			return nil, false, err
		}
		defer response.Body.Close()
		file = response.Body
		writeable = false

	} else {
		diskFile, err := os.Open(filename)
		if err != nil {
			return nil, false, err
		}
		defer diskFile.Close()
		file = diskFile
		writeable = true
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, false, err
	}

	contentType := http.DetectContentType(data)
	if contentType == "application/x-gzip" {
		writeable = false
		gz, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, false, err
		}
		defer gz.Close()
		data, err = io.ReadAll(gz)
		if err != nil {
			return nil, false, err
		}

	} else if contentType == "application/zip" {
		writeable = false
		z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, false, err
		}
		for _, zf := range z.File {
			f, err := zf.Open()
			if err != nil {
				return nil, false, err
			}
			defer f.Close()
			bytes, err := io.ReadAll(f)
			if err != nil {
				return nil, false, err
			}
			if storage.IsDiskette(bytes) {
				data = bytes
				break
			}
		}
	}

	return data, writeable, nil
}

// LoadDiskette returns a Diskette by detecting the format
func LoadDiskette(filename string) (storage.Diskette, error) {
	data, writeable, err := LoadResource(filename)
	if err != nil {
		return nil, err
	}

	return storage.MakeDiskette(data, filename, writeable)
}
