package vanmoof

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// External-SPI-flash secrets sector: 128 records × 32 B = 4 KB at 0x5A000.
// Each record is 28 B payload + 4 B CRC-32/LE (OEM crc32_le, seed
// 0xFFFFFFFF, no final XOR). See vanmoof/crc.c_ and secrets.c.
const (
	secretsSectorBase   int64  = 0x0005A000
	secretsRecordBytes  int    = 0x20
	secretsPayloadBytes int    = 0x1C
	secretsCRCSeed      uint32 = 0xFFFFFFFF
)

// CRCRecordSecret describes a slot that uses the 28+4 record-CRC layout.
type CRCRecordSecret struct {
	Name       string
	Slot       int
	PayloadLen int // meaningful bytes inside the 28 B payload
}

// RawSecret describes data read as-is, outside the record-CRC API.
type RawSecret struct {
	Name   string
	Offset int64
	Length int
}

func ReadSecrets(file *os.File) {
	crcSecrets := []CRCRecordSecret{
		{"BLE Authentication Key", 0, 16},
		{"Manufacturing Key", 126, 16},
	}

	for _, s := range crcSecrets {
		offset := secretsSectorBase + int64(s.Slot)*int64(secretsRecordBytes)
		record := readFromFile(file, offset, secretsRecordBytes)
		if len(record) < secretsRecordBytes {
			fmt.Printf("%s: short read (%d bytes)\n", s.Name, len(record))
			continue
		}

		calc := CalcCRC32LE(secretsCRCSeed, record[:secretsPayloadBytes])
		stored := binary.LittleEndian.Uint32(record[secretsPayloadBytes:])

		key := strings.ToUpper(hex.EncodeToString(record[:s.PayloadLen]))
		if calc == stored {
			fmt.Printf("%s: %s [CRC OK 0x%08X]\n", s.Name, key, stored)
		} else {
			fmt.Printf("%s: %s [CRC FAIL stored=0x%08X calc=0x%08X]\n",
				s.Name, key, stored, calc)
		}
	}

	// M-ID/M-KEY spans slots 124+125 as raw 60 B — no per-record CRC.
	rawSecrets := []RawSecret{
		{"M-ID/M-KEY", 0x0005AF80, 60},
	}
	for _, s := range rawSecrets {
		buf := readFromFile(file, s.Offset, s.Length)
		fmt.Printf("%s: %s\n", s.Name, hex.EncodeToString(buf))
	}

	// MAC address with MOOF validation
	macBuf := readFromFile(file, 0x0005AFE0, 16) // MAC (12) + MOOF (4)
	macStr := string(macBuf)
	if strings.HasSuffix(macStr, "MOOF") {
		fmt.Printf("MAC Address: %s\n", macStr[:12])
	} else {
		fmt.Println("MAC Address: MOOF signature not found")
	}
}
