package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	// Define the file paths
	readmePath := "../README.md"
	usagePath := "usage.txt"
	newUsagePath := "usage_new.txt"

	err := os.Rename(newUsagePath, usagePath)
	if err != nil {
		log.Fatal(err)
	}

	// Read the contents of the usage file
	usageBytes, err := ioutil.ReadFile(usagePath)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the usage bytes to string
	usage := string(usageBytes)

	// Read the contents of the readme file
	readmeBytes, err := ioutil.ReadFile(readmePath)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the readme bytes to string
	readme := string(readmeBytes)

	// Find the start and end markers
	startMarker := "<!-- doc/usage.txt start -->"
	endMarker := "<!-- doc/usage.txt end -->"
	startIndex := strings.Index(readme, startMarker)
	endIndex := strings.Index(readme, endMarker)

	// Replace the lines between start and end markers with the usage
	newReadme := readme[:startIndex+len(startMarker)] + "\n```terminal\n" + usage + "\n```\n" + readme[endIndex:]

	// Write the updated readme back to the file
	err = ioutil.WriteFile(readmePath, []byte(newReadme), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("README.md updated successfully!")
}
