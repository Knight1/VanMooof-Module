package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

func readSecrets(file os.File) {

	//
	//int64(0x0002000)

	// Offset in hex and length for the BLE authentication key
	offset := int64(0x005A000)
	length := 16 // Number of bytes to read

	buf := readFromFile(file, offset, length)

	// Convert to hex and print
	fmt.Println("BLE Authentication Key:", hex.EncodeToString(buf))
}

func readFromFile(file os.File, offset int64, length int) (buf []byte) {
	// Seek to the offset
	_, err := file.Seek(offset, 0)
	if err != nil {
		log.Fatalf("Failed to seek: %v", err)
	}

	// Read the bytes
	buf = make([]byte, length)
	_, err = file.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read: %v", err)
	}
	return buf

}
