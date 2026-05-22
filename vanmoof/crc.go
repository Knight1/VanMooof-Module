// crc32.go
// Go translation of the C utility for computing and verifying VanMoof ware binaries' CRC32

package vanmoof

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	wareMagic uint32 = 0xaa55aa55
	poly      uint32 = 0x04C11DB7
	initCRC   uint32 = 0xffffffff
	headSize         = 4*4 + 12 + 12 // magic, version, crc, length + date + time
)

type wareHeader struct {
	Magic   uint32
	Version uint32
	CRC     uint32
	Length  uint32
	Date    [12]byte
	Time    [12]byte
}

// polyLE is the reflected form of poly (0x04C11DB7), used by the OEM
// `crc32_le` in secrets.c (FUN_0002c726). Linux-kernel-style: init
// 0xFFFFFFFF, no final XOR.
const polyLE uint32 = 0xEDB88320

// Precomputed CRC tables for performance
var (
	crcTable   [256]uint32 // forward CRC-32 (ware binary format)
	crcTableLE [256]uint32 // reflected CRC-32 (OEM crc32_le)
)

func init() {
	for i := 0; i < 256; i++ {
		crc := uint32(i) << 24
		for j := 0; j < 8; j++ {
			if crc&(1<<31) != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
		crcTable[i] = crc

		c := uint32(i)
		for j := 0; j < 8; j++ {
			if c&1 != 0 {
				c = (c >> 1) ^ polyLE
			} else {
				c >>= 1
			}
		}
		crcTableLE[i] = c
	}
}

// CalcCRC32LE is the OEM `crc32_le` used by the external-flash secrets
// store (slot 0 BLE auth key, slot 126 manufacturing key, …). Matches
// FUN_0002c726 in the firmware: reflected CRC-32, seed taken from the
// caller, no final XOR. Pass 0xFFFFFFFF as seed for the secrets format.
func CalcCRC32LE(seed uint32, data []byte) uint32 {
	crc := seed
	for _, b := range data {
		crc = crcTableLE[(crc^uint32(b))&0xFF] ^ (crc >> 8)
	}
	return crc
}

// CalcCRC performs CRC32 over 32-bit little-endian words using lookup table
func CalcCRC(crc uint32, data []byte) uint32 {
	// Process 4-byte words
	for i := 0; i+4 <= len(data); i += 4 {
		word := binary.LittleEndian.Uint32(data[i : i+4])
		crc ^= word
		for j := 0; j < 4; j++ {
			crc = (crc << 8) ^ crcTable[(crc>>24)&0xff]
		}
	}
	
	// Handle remaining bytes (pad with zeros)
	remaining := len(data) % 4
	if remaining > 0 {
		var padded [4]byte
		copy(padded[:], data[len(data)-remaining:])
		word := binary.LittleEndian.Uint32(padded[:])
		crc ^= word
		for j := 0; j < 4; j++ {
			crc = (crc << 8) ^ crcTable[(crc>>24)&0xff]
		}
	}
	return crc
}

// VerifyWareFile verifies the CRC32 of a VanMoof ware file
func VerifyWareFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	size := len(data)
	if size < headSize {
		return fmt.Errorf("file too small: %d bytes, need at least %d", size, headSize)
	}

	// Parse header
	var h wareHeader
	head := make([]byte, headSize)
	copy(head, data[:headSize])
	h.Magic = binary.LittleEndian.Uint32(head[0:4])
	h.Version = binary.LittleEndian.Uint32(head[4:8])
	h.CRC = binary.LittleEndian.Uint32(head[8:12])
	h.Length = binary.LittleEndian.Uint32(head[12:16])
	copy(h.Date[:], head[16:28])
	copy(h.Time[:], head[28:40])

	crcVal := initCRC

	// Determine type by magic field
	if h.Magic == wareMagic {
		// Validate length
		if size < int(h.Length) {
			return fmt.Errorf("claimed length 0x%08x beyond file size 0x%08x", h.Length, size)
		}

		// Zero out CRC and length fields for calculation
		copy(head[8:12], []byte{0xff, 0xff, 0xff, 0xff})
		copy(head[12:16], []byte{0xff, 0xff, 0xff, 0xff})

		// Compute CRC over header + payload
		crcVal = CalcCRC(crcVal, head)
		crcVal = CalcCRC(crcVal, data[headSize:int(h.Length)])

		if crcVal != h.CRC {
			return fmt.Errorf("CRC mismatch: computed 0x%08x, expected 0x%08x", crcVal, h.CRC)
		}
	} else {
		// Extract expected CRC and version bytes
		if size < 8 {
			return fmt.Errorf("file too small for bootloader format: %d bytes", size)
		}
		expCRC := binary.LittleEndian.Uint32(data[size-4:])

		// Compute CRC over entire file except last 4 bytes
		crcVal = CalcCRC(crcVal, data[:size-4])

		if crcVal != expCRC {
			return fmt.Errorf("CRC mismatch: computed 0x%08x, expected 0x%08x", crcVal, expCRC)
		}
	}

	return nil
}
