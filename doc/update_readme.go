package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	// Define the file paths
	commandLinePath := "command_line.md"
	usagePath := "usage.txt"
	newUsagePath := "usage_new.txt"

	err := os.Rename(newUsagePath, usagePath)
	if err != nil {
		log.Fatal(err)
	}

	// Read the contents of the usage file
	usageBytes, err := os.ReadFile(usagePath)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the usage bytes to string
	usage := string(usageBytes)

	// Read the contents of the command_line.md file
	commandLineBytes, err := os.ReadFile(commandLinePath)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the command_line bytes to string
	commandLine := string(commandLineBytes)

	// Find the start and end markers
	startMarker := "<!-- doc/usage.txt start -->"
	endMarker := "<!-- doc/usage.txt end -->"
	startIndex := strings.Index(commandLine, startMarker)
	endIndex := strings.Index(commandLine, endMarker)

	// Replace the lines between start and end markers with the usage
	newCommandLine := commandLine[:startIndex+len(startMarker)] + "\n```terminal\n" + usage + "\n```\n" + commandLine[endIndex:]

	// Write the updated command_line.md back to the file
	err = os.WriteFile(commandLinePath, []byte(newCommandLine), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("command_line.md updated successfully!")
}
