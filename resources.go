package izapple2

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"embed"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func normalizeFilename(filename string) string {
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
	return filename
}

// LoadResource loads in memory a file from the filesystem, http or embedded
func LoadResource(filename string) ([]uint8, bool, error) {
	filename = normalizeFilename(filename)

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

/*
overlayFilename returns where to keep the changes a disk gets, or an empty
string when they go back to the image.

With a save directory, the changes the software writes go to an overlay file
there instead of to the image, which is then never modified. It also makes
writable the disks that could not be written at all, the ones loaded from a
compressed file, an URL or the embedded resources.
*/
func overlayFilename(filename string, saveDirectory string) string {
	if saveDirectory == "" {
		return ""
	}
	return filepath.Join(saveDirectory, filepath.Base(filename)+".ovl")
}

// LoadDiskette returns a Diskette by detecting the format. An empty
// saveDirectory writes the changes back to the image.
func LoadDiskette(filename string, saveDirectory string) (storage.Diskette, error) {
	data, writeable, err := LoadResource(filename)
	if err != nil {
		return nil, err
	}

	return storage.MakeDiskette(data, filename, writeable,
		overlayFilename(filename, saveDirectory))
}

// LoadDisketteFromBytes returns a Diskette from byte array (useful for WASM/browser file loading)
func LoadDisketteFromBytes(data []byte, filename string, writeable bool) (storage.Diskette, error) {
	return storage.MakeDiskette(data, filename, writeable, "")
}

// LoadBlockDisk returns a BlockDisk. An empty saveDirectory writes the changes
// back to the image.
func LoadBlockDisk(filename string, saveDirectory string) (storage.BlockDisk, error) {
	filename = normalizeFilename(filename)

	// Try to open as a file
	readOnly := false
	file, err := os.OpenFile(filename, os.O_RDWR, 0)
	if os.IsPermission(err) {
		// Retry in read-only mode
		readOnly = true
		file, _ = os.OpenFile(filename, os.O_RDONLY, 0)
	}

	var blockDisk storage.BlockDisk
	if file != nil {
		blockDisk, err = storage.NewBlockDiskFile(file, readOnly)
	} else {
		// Load as a resource
		var data []uint8
		data, _, err = LoadResource(filename)
		if err != nil {
			return nil, err
		}
		blockDisk, err = storage.NewBlockDiskMemory(data)
	}
	if err != nil {
		return nil, err
	}

	// With a save directory, the changes go to an overlay and the image is
	// left as it is, even the ones opened read only
	return storage.NewBlockDiskOverlay(blockDisk,
		overlayFilename(filename, saveDirectory))
}
