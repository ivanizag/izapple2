package storage

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ivanizag/izapple2/romdumps"
)

const (
	internalPrefix = "<internal>/"
	httpPrefix     = "http://"
	httpsPrefix    = "https://"
)

func isInternalResource(filename string) bool {
	return strings.HasPrefix(filename, internalPrefix)
}

func isHTTPResource(filename string) bool {
	return strings.HasPrefix(filename, httpPrefix) ||
		strings.HasPrefix(filename, httpsPrefix)
}

// LoadResource loads in memory a file from the filesystem, http or embedded
func LoadResource(filename string) ([]uint8, error) {
	var file io.Reader
	if isInternalResource(filename) {
		// load from embedded resource
		resource := strings.TrimPrefix(filename, internalPrefix)
		resourceFile, err := romdumps.Assets.Open(resource)
		if err != nil {
			return nil, err
		}
		defer resourceFile.Close()
		file = resourceFile

	} else if isHTTPResource(filename) {
		response, err := http.Get(filename)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		file = response.Body

	} else {
		diskFile, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer diskFile.Close()
		file = diskFile
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	contentType := http.DetectContentType(data)
	if contentType == "application/x-gzip" {
		gz, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		data, err = ioutil.ReadAll(gz)
		if err != nil {
			return nil, err
		}

	} else if contentType == "application/zip" {
		z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}
		for _, zf := range z.File {
			f, err := zf.Open()
			if err != nil {
				return nil, err
			}
			defer f.Close()
			bytes, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, err
			}
			if isFileDsk(bytes) || isFileNib(bytes) || isFileWoz(bytes) {
				data = bytes
				break
			}
		}
	}

	return data, nil
}
