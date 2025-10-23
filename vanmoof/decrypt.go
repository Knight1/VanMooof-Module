package vanmoof

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
)

// DecryptPack decrypts a VanMoof PACK file using AES ECB
func DecryptPack(packFile, keyHex string) error {
	// Read the encrypted pack file
	data, err := os.ReadFile(packFile)
	if err != nil {
		return fmt.Errorf("failed to read pack file: %v", err)
	}

	// Decode the hex key
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("invalid hex key: %v", err)
	}

	// Validate key length (AES-128 only - 16 bytes = 32 hex chars)
	if len(key) != 16 {
		return fmt.Errorf("invalid key length: %d bytes (expected 16 bytes / 32 hex characters)", len(key))
	}

	fmt.Printf("Decrypting %s with AES-128 ECB...\n", packFile)
	fmt.Printf("File size: %d bytes\n", len(data))

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %v", err)
	}

	// Check if data length is multiple of block size
	if len(data)%aes.BlockSize != 0 {
		return fmt.Errorf("encrypted data length (%d) is not a multiple of AES block size (%d)", len(data), aes.BlockSize)
	}

	// Decrypt using ECB mode
	decrypted := make([]byte, len(data))
	for i := 0; i < len(data); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], data[i:i+aes.BlockSize])
	}

	// Calculate CRC32 of decrypted data
	crc := crc32.ChecksumIEEE(decrypted)
	fmt.Printf("Decrypted data CRC32: 0x%08X\n", crc)

	// Generate output filename
	ext := filepath.Ext(packFile)
	base := packFile[:len(packFile)-len(ext)]
	outputFile := base + "_decrypted" + ext

	// Write decrypted data
	err = os.WriteFile(outputFile, decrypted, 0644)
	if err != nil {
		return fmt.Errorf("failed to write decrypted file: %v", err)
	}

	fmt.Printf("Decrypted PACK saved to: %s (%d bytes)\n", outputFile, len(decrypted))

	// Try to analyze the decrypted PACK
	if len(decrypted) >= 4 {
		magic := string(decrypted[0:4])
		if magic == "PACK" {
			fmt.Println("✓ Decryption successful - valid PACK magic found")
			analyzePack(decrypted)
		} else {
			fmt.Printf("⚠ Warning: No PACK magic found at start (found: %q)\n", magic)
			// Search for PACK magic in the decrypted data
			searchForPack(decrypted)
		}
	}

	return nil
}

// analyzePack analyzes the structure of a decrypted PACK file
func analyzePack(data []byte) {
	if len(data) < 12 {
		fmt.Println("PACK file too small for header")
		return
	}

	header := PackHeader{}
	copy(header.Magic[:], data[0:4])
	header.Offset = uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24
	header.Length = uint32(data[8]) | uint32(data[9])<<8 | uint32(data[10])<<16 | uint32(data[11])<<24

	fmt.Printf("PACK Header:\n")
	fmt.Printf("  Magic: %s\n", string(header.Magic[:]))
	fmt.Printf("  Directory Offset: 0x%08X (%d)\n", header.Offset, header.Offset)
	fmt.Printf("  Directory Length: 0x%08X (%d)\n", header.Length, header.Length)

	// Validate directory bounds
	dirOffset := int(header.Offset)
	dirLength := int(header.Length)

	if dirOffset+dirLength > len(data) {
		fmt.Printf("⚠ Warning: Directory extends beyond file bounds\n")
		return
	}

	// Calculate number of entries
	entrySize := 64 // 56 bytes filename + 4 bytes offset + 4 bytes length
	entryCount := dirLength / entrySize

	fmt.Printf("Directory contains %d entries:\n", entryCount)

	for i := 0; i < entryCount && i < 10; i++ { // Limit to first 10 entries
		entryOffset := dirOffset + (i * entrySize)
		if entryOffset+entrySize > len(data) {
			break
		}

		// Extract filename (first 56 bytes)
		filename := make([]byte, 56)
		copy(filename, data[entryOffset:entryOffset+56])

		// Remove null bytes
		var cleanName []byte
		for _, b := range filename {
			if b == 0 {
				break
			}
			cleanName = append(cleanName, b)
		}

		if len(cleanName) > 0 {
			// Extract offset and length
			fileOffset := uint32(data[entryOffset+56]) | uint32(data[entryOffset+57])<<8 |
				uint32(data[entryOffset+58])<<16 | uint32(data[entryOffset+59])<<24
			fileLength := uint32(data[entryOffset+60]) | uint32(data[entryOffset+61])<<8 |
				uint32(data[entryOffset+62])<<16 | uint32(data[entryOffset+63])<<24

			fmt.Printf("  %s (offset: 0x%08X, length: %d bytes)\n", string(cleanName), fileOffset, fileLength)
		}
	}

	if entryCount > 10 {
		fmt.Printf("  ... and %d more entries\n", entryCount-10)
	}
}

// searchForPack searches for PACK magic bytes in the data
func searchForPack(data []byte) {
	packMagic := []byte("PACK")
	for i := 0; i <= len(data)-4; i++ {
		if data[i] == packMagic[0] && data[i+1] == packMagic[1] &&
			data[i+2] == packMagic[2] && data[i+3] == packMagic[3] {
			fmt.Printf("Found PACK magic at offset 0x%08X\n", i)
			if i+12 <= len(data) {
				// Try to read header at this offset
				offset := uint32(data[i+4]) | uint32(data[i+5])<<8 | uint32(data[i+6])<<16 | uint32(data[i+7])<<24
				length := uint32(data[i+8]) | uint32(data[i+9])<<8 | uint32(data[i+10])<<16 | uint32(data[i+11])<<24
				fmt.Printf("  Directory Offset: 0x%08X, Length: 0x%08X\n", offset, length)
			}
		}
	}
}
