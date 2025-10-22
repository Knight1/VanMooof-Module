package vanmoof

import (
	"encoding/hex"
	"fmt"
	"os"
)

func ReadSecrets(file *os.File) {
	// Offset in hex and length for the BLE authentication key
	offset := int64(0x005A000)
	length := 16 // Number of bytes to read

	buf := readFromFile(file, offset, length)

	// Convert to hex and print
	fmt.Println("BLE Authentication Key:", hex.EncodeToString(buf))
}
