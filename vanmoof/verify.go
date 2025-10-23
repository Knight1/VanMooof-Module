package vanmoof

import (
	"fmt"
	"os"
	"sort"
)

// MemoryRegion represents a known region in the SPI flash
type MemoryRegion struct {
	Name   string
	Start  int
	End    int
	Length int
}

// VerifyDump checks if all data in the SPI dump is accounted for
func VerifyDump(moduleFileName string, showExtra bool) error {
	file, err := os.Open(moduleFileName)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %v", err)
	}
	fileSize := int(stat.Size())

	// Read entire file
	data, err := os.ReadFile(moduleFileName)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Define known regions
	var knownRegions []MemoryRegion

	// BLE Secrets
	knownRegions = append(knownRegions, MemoryRegion{
		Name:   "BLE Secrets",
		Start:  0x005A000,
		End:    0x005A000 + 60,
		Length: 60,
	})

	// M-ID/M-KEY
	knownRegions = append(knownRegions, MemoryRegion{
		Name:   "M-ID/M-KEY",
		Start:  0x005af80,
		End:    0x005af80 + 60,
		Length: 60,
	})

	// MAC Address + MOOF
	knownRegions = append(knownRegions, MemoryRegion{
		Name:   "MAC Address",
		Start:  0x0005AFE0,
		End:    0x0005AFE0 + 16,
		Length: 16,
	})

	// Find PACK region
	packOffset := 0x80000
	if len(data) > packOffset+12 {
		magicBytes := []byte{0x50, 0x41, 0x43, 0x4B} // "PACK"
		if len(data) >= packOffset+4 &&
			data[packOffset] == magicBytes[0] &&
			data[packOffset+1] == magicBytes[1] &&
			data[packOffset+2] == magicBytes[2] &&
			data[packOffset+3] == magicBytes[3] {

			// Calculate PACK size
			packSize := calculatePackSize(data, packOffset)
			knownRegions = append(knownRegions, MemoryRegion{
				Name:   "PACK Firmware",
				Start:  packOffset,
				End:    packOffset + packSize,
				Length: packSize,
			})
		}
	}

	// Find VM_SOUND files
	sounds := FindVMSounds(data)
	for i, sound := range sounds {
		knownRegions = append(knownRegions, MemoryRegion{
			Name:   fmt.Sprintf("VM_SOUND_%02d", i+1),
			Start:  sound.Offset,
			End:    sound.Offset + sound.Length,
			Length: sound.Length,
		})
	}

	// Logs region
	logOffset := 0x3fdd000
	if fileSize > logOffset {
		logSize := fileSize - logOffset
		knownRegions = append(knownRegions, MemoryRegion{
			Name:   "Logs",
			Start:  logOffset,
			End:    fileSize,
			Length: logSize,
		})
	}

	// Sort regions by start address
	sort.Slice(knownRegions, func(i, j int) bool {
		return knownRegions[i].Start < knownRegions[j].Start
	})

	// Calculate coverage
	totalKnown := 0
	for _, region := range knownRegions {
		totalKnown += region.Length
	}

	fmt.Printf("=== SPI Dump Verification ===\n")
	fmt.Printf("File size: %d bytes (0x%X)\n", fileSize, fileSize)
	fmt.Printf("Known regions: %d\n", len(knownRegions))
	fmt.Printf("Total known data: %d bytes (0x%X)\n", totalKnown, totalKnown)
	fmt.Printf("Coverage: %.2f%%\n", float64(totalKnown)/float64(fileSize)*100)

	fmt.Printf("\nKnown regions:\n")
	for _, region := range knownRegions {
		fmt.Printf("  %-20s 0x%08X - 0x%08X (%d bytes)\n",
			region.Name, region.Start, region.End, region.Length)
	}

	if showExtra {
		fmt.Printf("\nUnaccounted regions:\n")
		findUnaccountedRegions(knownRegions, fileSize, data)
	}

	return nil
}

// calculatePackSize determines the actual size of the PACK data
func calculatePackSize(data []byte, offset int) int {
	if len(data) < offset+12 {
		return 0
	}

	// Read directory offset and length from PACK header
	dirOffset := int(data[offset+4]) | int(data[offset+5])<<8 | int(data[offset+6])<<16 | int(data[offset+7])<<24
	dirLength := int(data[offset+8]) | int(data[offset+9])<<8 | int(data[offset+10])<<16 | int(data[offset+11])<<24

	// PACK size = header (12) + data + directory
	return 12 + dirOffset + dirLength
}

// findUnaccountedRegions identifies gaps between known regions
func findUnaccountedRegions(knownRegions []MemoryRegion, fileSize int, data []byte) {
	if len(knownRegions) == 0 {
		fmt.Printf("  0x%08X - 0x%08X (%d bytes) - Entire file unaccounted\n",
			0, fileSize, fileSize)
		return
	}

	// Check for gap at the beginning
	if knownRegions[0].Start > 0 {
		size := knownRegions[0].Start
		if isSignificantRegion(data, 0, size) {
			fmt.Printf("  0x%08X - 0x%08X (%d bytes) - Unknown data\n",
				0, knownRegions[0].Start, size)
		}
	}

	// Check for gaps between regions
	for i := 0; i < len(knownRegions)-1; i++ {
		currentEnd := knownRegions[i].End
		nextStart := knownRegions[i+1].Start

		if nextStart > currentEnd {
			size := nextStart - currentEnd
			if isSignificantRegion(data, currentEnd, size) {
				fmt.Printf("  0x%08X - 0x%08X (%d bytes) - Unknown data\n",
					currentEnd, nextStart, size)
			}
		}
	}

	// Check for gap at the end
	lastRegion := knownRegions[len(knownRegions)-1]
	if lastRegion.End < fileSize {
		size := fileSize - lastRegion.End
		if isSignificantRegion(data, lastRegion.End, size) {
			fmt.Printf("  0x%08X - 0x%08X (%d bytes) - Unknown data\n",
				lastRegion.End, fileSize, size)
		}
	}
}

// isSignificantRegion checks if a region contains non-zero/non-FF data
func isSignificantRegion(data []byte, offset, size int) bool {
	if offset+size > len(data) {
		size = len(data) - offset
	}
	if size <= 0 {
		return false
	}

	// Skip very small regions (less than 16 bytes)
	if size < 16 {
		return false
	}

	// Check if region contains significant data (not all 0x00 or 0xFF)
	zeroCount := 0
	ffCount := 0

	for i := offset; i < offset+size && i < len(data); i++ {
		if data[i] == 0x00 {
			zeroCount++
		} else if data[i] == 0xFF {
			ffCount++
		}
	}

	// Consider significant if less than 95% is padding bytes
	paddingRatio := float64(zeroCount+ffCount) / float64(size)
	return paddingRatio < 0.95
}
