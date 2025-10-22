package vanmoof

import (
	"encoding/hex"
	"fmt"
	"os"
)

func readFromFile(file *os.File, offset int64, length int) []byte {
	buf := make([]byte, length)
	_, err := file.ReadAt(buf, offset)
	if err != nil {
		fmt.Printf("Error reading from file: %v\n", err)
		return nil
	}
	return buf
}

func ReadLogs(file *os.File) {
	offset := int64(0x3fdd000)
	length := 16 // Number of bytes to read

	buf := readFromFile(file, offset, length)

	// Convert to hex and print
	fmt.Println("Logs:", hex.EncodeToString(buf))
}
