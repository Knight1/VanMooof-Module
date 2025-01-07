package main

import (
	"encoding/hex"
	"fmt"
	"os"
)

func readLogs(file os.File) {

	offset := int64(0x3fdd000)
	length := 16 // Number of bytes to read

	buf := readFromFile(file, offset, length)

	// Convert to hex and print
	fmt.Println("Logs:", hex.EncodeToString(buf))
}
