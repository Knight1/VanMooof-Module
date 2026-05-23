// crc.go — CRC-32 routines used across the VanMoof firmware ecosystem.
//
// Two distinct CRC-32 variants coexist in the OEM stack and both are
// implemented here:
//
//  1. Forward CRC-32 (poly 0x04C11DB7, non-reflected, init 0xFFFFFFFF,
//     no final XOR). Used to authenticate VanMoof "ware" binaries
//     (bootloader/firmware images). See CalcCRC and VerifyWareFile.
//
//  2. Reflected CRC-32 / Linux-kernel `crc32_le` (poly 0xEDB88320,
//     init 0xFFFFFFFF, no final XOR). Used by the external-SPI-flash
//     secrets sector at 0x5A000 to protect each 32-byte record's
//     28-byte payload. See CalcCRC32LE; readSecrets.go consumes it.
//
// Both tables are precomputed in init().

package vanmoof

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	// wareMagic identifies a VanMoof "ware" image (firmware/bootloader)
	// in the first 4 bytes of its header. Files without this magic are
	// treated as raw images whose CRC lives in the trailing 4 bytes.
	wareMagic uint32 = 0xaa55aa55

	// poly is the IEEE 802.3 CRC-32 polynomial in non-reflected form,
	// used by the ware-binary CRC (see CalcCRC).
	poly uint32 = 0x04C11DB7

	// initCRC is the standard CRC-32 seed shared by both variants.
	initCRC uint32 = 0xffffffff

	// headSize is the on-disk size of wareHeader: 4 × uint32 (magic,
	// version, crc, length) followed by 12-byte date and 12-byte time
	// strings.
	headSize = 4*4 + 12 + 12
)

// wareHeader is the fixed-layout prefix of a VanMoof ware binary. The
// CRC field covers the header (with CRC and Length blanked to 0xFF)
// plus the payload up to Length bytes.
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

// Precomputed 256-entry lookup tables, populated once in init():
//
//   - crcTable:   forward CRC-32 (poly 0x04C11DB7, MSB-first), used by
//     CalcCRC for the ware-binary format.
//   - crcTableLE: reflected CRC-32 (poly 0xEDB88320, LSB-first), used
//     by CalcCRC32LE to match the OEM crc32_le helper.
var (
	crcTable   [256]uint32
	crcTableLE [256]uint32
)

// init builds both CRC-32 lookup tables. For each byte value 0..255 it
// runs 8 polynomial-division steps twice: once MSB-first with poly to
// fill crcTable (forward CRC), and once LSB-first with polyLE to fill
// crcTableLE (reflected CRC). Done once at package load so the runtime
// path is a single table indexing per byte.
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

// CalcCRC32LE computes the reflected CRC-32 used by the OEM `crc32_le`
// helper in the external-flash secrets store (slot 0 BLE auth key,
// slot 126 manufacturing key, ...). It mirrors firmware function
// FUN_0002c726:
//
//	crc = table[(crc ^ byte) & 0xFF] ^ (crc >> 8)
//
// The caller supplies the seed (use 0xFFFFFFFF for the secrets-record
// format) and no final XOR is applied — the returned value is the raw
// register state, matching what the firmware stores at byte offset
// 0x1C of each 32-byte record.
func CalcCRC32LE(seed uint32, data []byte) uint32 {
	crc := seed
	for _, b := range data {
		crc = crcTableLE[(crc^uint32(b))&0xFF] ^ (crc >> 8)
	}
	return crc
}

// CalcCRC computes the forward CRC-32 used by VanMoof ware binaries.
// It consumes data as 32-bit little-endian words: each word is XORed
// into the running register, then advanced 4 bytes through the
// MSB-first table (crcTable, poly 0x04C11DB7). A trailing partial word
// is zero-padded to 4 bytes so the same word-oriented step applies.
//
// Use initCRC (0xFFFFFFFF) for `crc` on the first call. No final XOR
// is applied; the returned register state is compared directly against
// the value stored in the binary's header (or trailing 4 bytes for
// bootloader-format files).
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

// VerifyWareFile validates the CRC-32 of a VanMoof ware file on disk.
//
// Two on-disk formats are accepted, distinguished by the first 4 bytes:
//
//   - wareHeader format (magic == wareMagic): the header's CRC field
//     covers the header (with the CRC and Length words blanked to
//     0xFF) plus the payload up to header.Length bytes. The function
//     reads the header, recomputes the CRC with CalcCRC, and compares
//     it to header.CRC.
//
//   - Bootloader/raw format (any other magic): the CRC sits in the
//     last 4 little-endian bytes of the file and covers everything
//     before it. The function recomputes the CRC over data[:size-4]
//     and compares it to that trailing word.
//
// Returns nil on a match, or a descriptive error on mismatch, short
// reads, or claimed lengths that exceed the file size.
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
