package vanmoof

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

type Secret struct {
	Name   string
	Offset int64
	Length int
}

func ReadSecrets(file *os.File) {
	secrets := []Secret{
		{"BLE Authentication Key", 0x005A000, 16},
		{"Manufacturing Key", 0x005AFC0, 16},
		{"M-ID/M-KEY", 0x005af80, 60},
		// Add more secrets here as needed
	}

	for _, secret := range secrets {
		buf := readFromFile(file, secret.Offset, secret.Length)
		if secret.Name == "BLE Authentication Key" || secret.Name == "Manufacturing Key" {
			fmt.Printf("%s: %s\n", secret.Name, strings.ToUpper(hex.EncodeToString(buf)))
		} else {
			fmt.Printf("%s: %s\n", secret.Name, hex.EncodeToString(buf))
		}
	}

	// Read MAC address with MOOF validation
	macBuf := readFromFile(file, 0x0005AFE0, 16) // MAC (12) + MOOF (4)
	macStr := string(macBuf)
	if strings.HasSuffix(macStr, "MOOF") {
		mac := macStr[:12] // Extract MAC address part
		fmt.Printf("MAC Address: %s\n", mac)
	} else {
		fmt.Println("MAC Address: MOOF signature not found")
	}
}
