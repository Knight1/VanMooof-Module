package vanmoof

import (
	"fmt"
	"os"
)

const (
	readCommand = 0x03             // Low-speed read command
	flashSize   = 64 * 1024 * 1024 // 64 MB
	chunkSize   = 256              // 256 bytes per read operation
)

// DumpFlash reads the entire SPI flash chip and saves it to a file
func DumpFlash(macAddress, frameNumber string) error {
	conn, err := spiConnect()
	if err != nil {
		return fmt.Errorf("failed to connect to SPI: %v", err)
	}

	// Open output file
	file, err := os.Create("VMES3-" + frameNumber + "-" + macAddress + ".bin")
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
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
			return fmt.Errorf("SPI transaction failed at offset 0x%X: %v", offset, err)
		}

		// Write the data to the file
		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("failed to write data to file at offset 0x%X: %v", offset, err)
		}

		fmt.Printf("Read and saved chunk at offset 0x%X\n", offset)
	}

	return nil
}
