package main

import (
	"fmt"
	"os"
)

const (
	readCommand = 0x03             // Low-speed read command
	flashSize   = 64 * 1024 * 1024 // 64 MB
	chunkSize   = 256              // 256 bytes per read operation
)

func dumpFlash() {
	// MAC
	// FRAME

	macAddress := ""
	frameNumber := ""

	conn, err := spiConnect()
	if err != nil {
		fmt.Println("")
	}

	// Open output file
	file, err := os.Create("VMES3-" + frameNumber + "-" + macAddress + ".bin")
	if err != nil {
		fmt.Printf("Failed to create output file: %v", err)
		return
	}
	defer file.Close()

	// Sequentially read 64MB in 256-byte chunks
	for offset := 0; offset < flashSize; offset += chunkSize {
		// Set up 24-bit address
		address := []byte{
			byte(offset >> 16),
			byte(offset >> 8),
			byte(offset),
		}

		// Prepare the read command with the address
		readBuffer := append([]byte{readCommand}, address...)

		// Buffer to store incoming data
		data := make([]byte, chunkSize)

		// Execute SPI read
		if err := conn.Tx(readBuffer, data); err != nil {
			fmt.Printf("SPI transaction failed at offset 0x%X: %v", offset, err)
			return
		}

		// Write the data to the file
		if _, err := file.Write(data); err != nil {
			fmt.Printf("Failed to write data to file at offset 0x%X: %v", offset, err)
			return
		}

		fmt.Printf("Read and saved chunk at offset 0x%X\n", offset)
	}
}
