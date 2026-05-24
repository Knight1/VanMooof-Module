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

	// Secrets sector (0x5A000, 4 KB) — 128 × 32-byte CRC-protected
	// records. Layout derived from bleware/src/secrets.c and auth.c.
	// We label one region per slot group so every byte of the sector
	// is accounted for in the -verify output.
	const (
		secretsBase    = 0x0005A000
		recordBytes    = 0x20
		uKeyFirstSlot  = 1   // slots 1..123 are the user-keyed range
		uKeyLastSlot   = 123 // (slot 0 holds the BLE auth key)
		midSlot        = 124
		reservedSlot   = 125
		mfgSlot        = 126
		macSlot        = 127
	)
	slotAddr := func(slot int) int { return secretsBase + slot*recordBytes }

	knownRegions = append(knownRegions,
		MemoryRegion{
			Name:   "BLE Auth Key (slot 0)",
			Start:  slotAddr(0),
			End:    slotAddr(1),
			Length: recordBytes,
		},
		MemoryRegion{
			Name:   "UKEY Records (slots 1-123)",
			Start:  slotAddr(uKeyFirstSlot),
			End:    slotAddr(uKeyLastSlot + 1),
			Length: (uKeyLastSlot - uKeyFirstSlot + 1) * recordBytes,
		},
		MemoryRegion{
			Name:   "M-ID Record (slot 124)",
			Start:  slotAddr(midSlot),
			End:    slotAddr(midSlot + 1),
			Length: recordBytes,
		},
		MemoryRegion{
			Name:   "Reserved (slot 125)",
			Start:  slotAddr(reservedSlot),
			End:    slotAddr(reservedSlot + 1),
			Length: recordBytes,
		},
		MemoryRegion{
			Name:   "Manufacturing Key (slot 126)",
			Start:  slotAddr(mfgSlot),
			End:    slotAddr(mfgSlot + 1),
			Length: recordBytes,
		},
		MemoryRegion{
			Name:   "MAC Address (slot 127)",
			Start:  slotAddr(macSlot),
			End:    slotAddr(macSlot + 1),
			Length: recordBytes,
		},
	)

	// BLEBoot OAD staging area — the BIM walks 44 candidate slots at
	// a 4 KB stride (see bleboot/src/oad.c — bim_full_scan_and_launch),
	// so the full range 0x0000..0x2C000 is reserved for OAD images
	// regardless of which slots happen to be populated. Always mark
	// the staging area as known so it doesn't bleed into the
	// unaccounted-regions output; populated slots get reported with
	// their version + status separately via PrintBLEBootImage.
	bleStagingLen := BLEBootSlotCount * int(BLEBootSlotStride)
	knownRegions = append(knownRegions, MemoryRegion{
		Name:   fmt.Sprintf("BLEBoot OAD Staging (slots 0-%d)", BLEBootSlotCount-1),
		Start:  0,
		End:    bleStagingLen,
		Length: bleStagingLen,
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

	// Find VM_SOUND files. Slot indices match the firmware's
	// `audio_play <index>` argument (0..0x7A inclusive).
	sounds := FindVMSounds(data)
	for _, sound := range sounds {
		knownRegions = append(knownRegions, MemoryRegion{
			Name:   fmt.Sprintf("VM_SOUND slot %d", sound.Slot),
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

	zeroPct, ffPct := paddingComposition(data, 0, fileSize)
	otherPct := 100 - zeroPct - ffPct

	fmt.Printf("=== SPI Dump Verification ===\n")
	fmt.Printf("File size: %d bytes (0x%X)\n", fileSize, fileSize)
	fmt.Printf("Known regions: %d\n", len(knownRegions))
	fmt.Printf("Total known data: %d bytes (0x%X)\n", totalKnown, totalKnown)
	fmt.Printf("Coverage: %.2f%%\n", float64(totalKnown)/float64(fileSize)*100)
	fmt.Printf("Byte composition: zeros %.2f%%, FFs %.2f%%, other %.2f%%\n",
		zeroPct, ffPct, otherPct)

	PrintBLEBootImage(file, fileSize)

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
		reportUnaccounted(data, 0, knownRegions[0].Start)
	}

	// Check for gaps between regions
	for i := 0; i < len(knownRegions)-1; i++ {
		currentEnd := knownRegions[i].End
		nextStart := knownRegions[i+1].Start
		if nextStart > currentEnd {
			reportUnaccounted(data, currentEnd, nextStart)
		}
	}

	// Check for gap at the end
	lastRegion := knownRegions[len(knownRegions)-1]
	if lastRegion.End < fileSize {
		reportUnaccounted(data, lastRegion.End, fileSize)
	}
}

// reportUnaccounted prints one [start, end) gap when its contents are
// substantive (see isSignificantRegion). The "Unknown data" line is
// suffixed with the zero/FF ratios so it's obvious whether what
// survived the filter is structured data or trailing padding.
func reportUnaccounted(data []byte, start, end int) {
	size := end - start
	if !isSignificantRegion(data, start, size) {
		return
	}
	zeroPct, ffPct := paddingComposition(data, start, size)
	fmt.Printf("  0x%08X - 0x%08X (%d bytes) - Unknown data (zeros %.1f%%, FFs %.1f%%)\n",
		start, end, size, zeroPct, ffPct)
}

// paddingComposition returns the percentage of 0x00 and 0xFF bytes
// in data[offset:offset+size], clamped to len(data).
func paddingComposition(data []byte, offset, size int) (zeroPct, ffPct float64) {
	if offset+size > len(data) {
		size = len(data) - offset
	}
	if size <= 0 {
		return 0, 0
	}
	zeroCount := 0
	ffCount := 0
	for i := offset; i < offset+size && i < len(data); i++ {
		switch data[i] {
		case 0x00:
			zeroCount++
		case 0xFF:
			ffCount++
		}
	}
	return float64(zeroCount) * 100 / float64(size),
		float64(ffCount) * 100 / float64(size)
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

	// Count padding bytes (0x00 and 0xFF) in a single pass.
	zeroCount := 0
	ffCount := 0
	for i := offset; i < offset+size && i < len(data); i++ {
		switch data[i] {
		case 0x00:
			zeroCount++
		case 0xFF:
			ffCount++
		}
	}

	// Filter out regions that are essentially all 0x00 — those are
	// uninitialised flash and tell us nothing about the on-disk
	// layout. The threshold is intentionally tight (99.9%) so that
	// any actual structure embedded in a mostly-zero region still
	// surfaces.
	if float64(zeroCount)/float64(size) >= 0.999 {
		return false
	}

	// Same idea for 0xFF (erased NOR flash).
	if float64(ffCount)/float64(size) >= 0.999 {
		return false
	}

	// Consider significant if less than 95% is mixed padding bytes.
	paddingRatio := float64(zeroCount+ffCount) / float64(size)
	return paddingRatio < 0.95
}
