package vanmoof

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CheckForFirmware(moduleFileName *string) {
	if *moduleFileName == "" {
		return
	}

	file, err := os.Open(*moduleFileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Read entire file into memory
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Check PACK at fixed offset 0x80000
	offset := 0x80000
	if len(data) < offset+4 {
		fmt.Println("File too small to contain PACK at 0x80000")
		return
	}

	// Check if PACK magic exists at fixed location
	magicBytes := []byte{0x50, 0x41, 0x43, 0x4B} // "PACK"
	if !bytes.Equal(data[offset:offset+4], magicBytes) {
		fmt.Println("No PACK firmware found at 0x80000")
		return
	}

	fmt.Printf("Found PACK at offset: 0x%08X\n", offset)

	// Read PACK header
	if len(data) < offset+12 {
		fmt.Println("Insufficient data for PACK header")
		return
	}

	header := PackHeader{}
	copy(header.Magic[:], data[offset:offset+4])
	header.Offset = binary.LittleEndian.Uint32(data[offset+4 : offset+8])
	header.Length = binary.LittleEndian.Uint32(data[offset+8 : offset+12])

	fmt.Printf("PACK Header - Offset: 0x%08X, Length: 0x%08X\n", header.Offset, header.Length)

	// Calculate total PACK size (12 byte header + data up to directory end)
	totalPackSize := 12 + int(header.Offset) + int(header.Length)
	if offset+totalPackSize > len(data) {
		totalPackSize = len(data) - offset
	}

	// Extract PACK file with correct length
	packData := data[offset : offset+totalPackSize]
	packFileName := filepath.Base(*moduleFileName) + ".pack"

	err = os.WriteFile(packFileName, packData, 0644)
	if err != nil {
		fmt.Printf("Error writing PACK file: %v\n", err)
		return
	}

	fmt.Printf("Extracted PACK to: %s (%d bytes)\n", packFileName, len(packData))

	// Extract individual firmware files from PACK
	extractPACK(packData, offset)
}

func extractPACK(packData []byte, baseOffset int) {
	if len(packData) < 12 {
		return
	}

	header := PackHeader{}
	copy(header.Magic[:], packData[0:4])
	header.Offset = binary.LittleEndian.Uint32(packData[4:8])
	header.Length = binary.LittleEndian.Uint32(packData[8:12])

	dirOffset := int(header.Offset)
	dirLength := int(header.Length)

	if dirOffset+dirLength > len(packData) {
		fmt.Printf("Directory beyond PACK data bounds\n")
		return
	}

	// Read directory entries
	entrySize := 64 // 56 bytes filename + 4 bytes offset + 4 bytes length
	entryCount := dirLength / entrySize

	fmt.Printf("Extracting %d firmware files:\n", entryCount)

	for i := 0; i < entryCount; i++ {
		entryOffset := dirOffset + (i * entrySize)
		if entryOffset+entrySize > len(packData) {
			break
		}

		// Read entry
		entry := PackEntry{}
		copy(entry.Filename[:], packData[entryOffset:entryOffset+56])
		entry.Offset = binary.LittleEndian.Uint32(packData[entryOffset+56 : entryOffset+60])
		entry.Length = binary.LittleEndian.Uint32(packData[entryOffset+60 : entryOffset+64])

		// Extract filename (remove null bytes)
		filename := string(bytes.TrimRight(entry.Filename[:], "\x00"))
		if filename == "" {
			continue
		}

		// Validate data bounds
		dataStart := int(entry.Offset)
		dataEnd := dataStart + int(entry.Length)
		if dataEnd > dirOffset {
			fmt.Printf("Skipping %s: data overruns directory\n", filename)
			continue
		}

		fmt.Printf("  %s (%d bytes)\n", filename, entry.Length)

		// Extract file data
		fileData := packData[dataStart:dataEnd]
		err := os.WriteFile(filename, fileData, 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
		}
	}
}

type PackHeader struct {
	Magic  [4]byte
	Offset uint32
	Length uint32
}

type PackEntry struct {
	Filename [56]byte
	Offset   uint32
	Length   uint32
}
