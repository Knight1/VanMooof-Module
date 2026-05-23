// readSecrets.go — decode the external-SPI-flash "secrets" sector.
//
// At external-flash offset 0x5A000 the VanMoof firmware keeps a 4 KB
// sector laid out as 128 records of 32 bytes each:
//
//	record[0..27]   payload (28 bytes)
//	record[28..31]  CRC-32/LE of payload, little-endian
//
// The CRC variant is the OEM `crc32_le` (reflected, seed 0xFFFFFFFF,
// no final XOR) implemented by CalcCRC32LE in crc.go. Known slots:
//
//	slot 0    BLE Authentication Key (16 B + 12 B pad, CRC protected)
//	slot 124  M-ID/M-KEY part 1  ┐ raw 60 B span (124+125), NOT
//	slot 125  M-ID/M-KEY part 2  ┘ wrapped in the record-CRC format
//	slot 126  Manufacturing Key       (16 B + 12 B pad, CRC protected)
//
// References: secrets.c (OEM), vanmoof/crc.c_, firmware functions
// secrets_record_read @ 0x00020BB8 / secrets_record_write_verify
// @ 0x00020C06.

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
// 0xFFFFFFFF, no final XOR).
const (
	// secretsSectorBase is the external-flash byte offset of the
	// 4 KB secrets sector.
	secretsSectorBase int64 = 0x0005A000

	// secretsRecordBytes is the total size of one record (payload +
	// trailing CRC).
	secretsRecordBytes int = 0x20

	// secretsPayloadBytes is the size of the CRC-covered prefix of a
	// record; the trailing 4 bytes hold the stored CRC.
	secretsPayloadBytes int = 0x1C

	// secretsCRCSeed is the initial CRC register value used by the
	// OEM crc32_le helper for these records.
	secretsCRCSeed uint32 = 0xFFFFFFFF
)

// CRCRecordSecret names a slot whose 32-byte record follows the
// 28 B payload + 4 B CRC layout. PayloadLen indicates how many of the
// 28 payload bytes are meaningful (the rest is padding).
type CRCRecordSecret struct {
	Name       string
	Slot       int
	PayloadLen int
}

// RawSecret names a region of the secrets sector that is read as-is,
// bypassing the per-record CRC API (e.g. the M-ID/M-KEY blob).
type RawSecret struct {
	Name   string
	Offset int64
	Length int
}

// ReadSecrets dumps every known entry from the external-flash secrets
// sector of the provided dump file. CRC-protected slots are verified
// against their stored CRC (printed as "CRC OK" / "CRC FAIL"); raw
// slots are dumped verbatim. Finally the bike's MAC address is read
// from 0x5AFE0 and validated by checking for the trailing "MOOF"
// signature. Results are written to stdout.
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
