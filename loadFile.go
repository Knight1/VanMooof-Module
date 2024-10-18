package main

import (
	"fmt"
	"log"
	"os"
)

func loadFile() (file *os.File) {
	// File path
	filePath := moduleFileName

	if *filePath == "" {
		fmt.Println("File path required")
		os.Exit(1)
	}

	fmt.Println("Loading File:", string(*filePath))

	// Open the file
	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}

	return file
}
