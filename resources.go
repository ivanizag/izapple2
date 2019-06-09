package apple2

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ivanizag/apple2/romdumps"
)

const (
	internalPrefix = "<internal>/"
)

func isInternalResource(filename string) bool {
	return strings.HasPrefix(filename, internalPrefix)
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
