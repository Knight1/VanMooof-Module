package vanmoof

import (
	"fmt"
	"log"
	"os"
)

func LoadFile(moduleFileName *string) (file *os.File) {
	// File path
	filePath := moduleFileName

	if *filePath == "" {
		fmt.Println("File path required. Use -f FILE")
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
