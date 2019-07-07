package apple2

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ivanizag/apple2/romdumps"
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

func loadResource(filename string) []uint8 {
	var file io.Reader
	if isInternalResource(filename) {
		// load from embedded resource
		resource := strings.TrimPrefix(filename, internalPrefix)
		resourceFile, err := romdumps.Assets.Open(resource)
		if err != nil {
			panic(err)
		}
		defer resourceFile.Close()
		file = resourceFile

	} else if isHTTPResource(filename) {
		response, err := http.Get(filename)
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()
		file = response.Body

	} else {
		diskFile, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		defer diskFile.Close()
		file = diskFile
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	return data
}
