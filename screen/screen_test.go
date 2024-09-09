package screen

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestSnapshots(t *testing.T) {
	// Verifies all the scenarios on the ./test_resources folder
	files, err := os.ReadDir("./test_resources/")
	if err != nil {
		t.Fatal(err)
	}
	for _, fileInfo := range files {
		if !fileInfo.IsDir() &&
			strings.HasSuffix(fileInfo.Name(), ".json") {
			testScenario(t, "./test_resources/"+fileInfo.Name())
		}
	}
}

func testScenario(t *testing.T, fileName string) {
	ts, err := loadTestScenario(fileName)
	if err != nil {
		t.Fatal(err)
	}

	ts.generateSnapshots(fileName, true)

	for _, screenMode := range ts.ScreenModes {
		referenceName := buildImageName(fileName, screenMode, false)
		actualName := buildImageName(fileName, screenMode, true)

		reference, err := os.ReadFile(referenceName)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := os.ReadFile(actualName)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(reference, actual) {
			t.Errorf("Files %s and %s should be equal", referenceName, actualName)
			replaceIfNeeded(referenceName, actualName)
		} else {
			os.Remove(actualName)
		}

	}
}

func replaceIfNeeded(referenceName string, actualName string) {
	// If the "update" argument is passed to test. The new images replace the old.
	// Run the tests with: "go test . -args update"
	for _, arg := range os.Args {
		if arg == "update" {
			os.Remove(referenceName)
			os.Rename(actualName, referenceName)
		}
	}
}
