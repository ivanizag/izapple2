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

func loadResource(filename string) []uint8 {
	var file io.Reader
	if strings.HasPrefix(filename, internalPrefix) {
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
