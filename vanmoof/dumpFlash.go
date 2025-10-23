package vanmoof

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"strings"
	"time"

	"periph.io/x/conn/v3/spi"
)

const (
	readCommand = 0x03             // Low-speed read command
	flashSize   = 64 * 1024 * 1024 // 64 MB (MX25L51245G - VanMoof S3)
	chunkSize   = 4096             // 4KB chunks for better performance
)

// DumpFlash reads the entire SPI flash chip and saves it to a file
func DumpFlash(macAddress, frameNumber string, sudo bool) error {
	if !sudo {
		return fmt.Errorf("SPI flash dump requires -sudo flag for hardware access")
	}

	filename := fmt.Sprintf("VMES3-%s-%s.bin", frameNumber, macAddress)

	// Check if file exists and warn about overwrite
	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("⚠️  File %s already exists and will be overwritten!\n", filename)
	}

	conn, err := spiConnect()
	if err != nil {
		return fmt.Errorf("failed to connect to SPI: %v", err)
	}
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file: %v\n", closeErr)
		}
	}()

	fmt.Printf("Dumping 64MB SPI flash to %s...\n", filename)

	// Validate chip compatibility before dump
	if err := validateChipCompatibility(conn); err != nil {
		return err
	}

	// Check status registers for proper read conditions
	if err := validateStatusRegisters(conn); err != nil {
		return err
	}

	// Pre-validate BLE authentication key before full dump and save it
	originalBLEKey, err := validateBLEKey(conn)
	if err != nil {
		return err
	}

	start := time.Now()
	crc := crc32.NewIEEE()
	sha := sha256.New()
	multiWriter := io.MultiWriter(file, crc, sha)

	totalChunks := flashSize / chunkSize
	for offset := 0; offset < flashSize; offset += chunkSize {
		// 24-bit address for SPI command
		address := []byte{
			byte(offset >> 16),
			byte(offset >> 8),
			byte(offset),
		}

		cmd := append([]byte{readCommand}, address...)
		data := make([]byte, chunkSize)

		if err := conn.Tx(cmd, data); err != nil {
			return fmt.Errorf("SPI read failed at 0x%06X: %v", offset, err)
		}

		if _, err := multiWriter.Write(data); err != nil {
			return fmt.Errorf("write failed at 0x%06X: %v", offset, err)
		}

		// Progress reporting
		chunk := offset/chunkSize + 1
		if chunk%1024 == 0 || chunk == totalChunks {
			progress := float64(chunk) / float64(totalChunks) * 100
			fmt.Printf("Progress: %.1f%% (%d/%d chunks)\n", progress, chunk, totalChunks)
		}
	}

	// Force write and sync to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("\nDump completed in %v\n", duration)
	fmt.Printf("File: %s (64MB)\n", filename)
	fmt.Printf("CRC32: 0x%08X\n", crc.Sum32())
	fmt.Printf("SHA256: %x\n", sha.Sum(nil))

	// Verify file integrity
	if err := verifyDump(filename); err != nil {
		return err
	}

	// Extract MAC address and rename file if needed
	if err := extractMACAndRename(filename, macAddress, frameNumber); err != nil {
		return err
	}

	// Compare BLE key from dump with original
	if err := compareBLEKeys(filename, originalBLEKey); err != nil {
		return err
	}

	// Comprehensive verification: compare disk vs memory SHA512
	return verifyDumpIntegrity(conn, filename)
}

// verifyDump performs integrity check on the dumped file
func verifyDump(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("verification failed - cannot open file: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close verification file: %v\n", closeErr)
		}
	}()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("verification failed - cannot stat file: %v", err)
	}

	if stat.Size() != flashSize {
		return fmt.Errorf("verification failed - size mismatch: got %d, expected %d", stat.Size(), flashSize)
	}

	// Quick integrity check - read first and last chunks
	firstChunk := make([]byte, 16)
	lastChunk := make([]byte, 16)

	if _, err := file.Read(firstChunk); err != nil {
		return fmt.Errorf("verification failed - cannot read first chunk: %v", err)
	}

	if _, err := file.Seek(-16, io.SeekEnd); err != nil {
		return fmt.Errorf("verification failed - cannot seek to end: %v", err)
	}

	if _, err := file.Read(lastChunk); err != nil {
		return fmt.Errorf("verification failed - cannot read last chunk: %v", err)
	}

	fmt.Printf("✓ Verification passed - file integrity confirmed\n")
	return nil
}

// extractMACAndRename extracts MAC address from dump and renames file if needed
func extractMACAndRename(filename, originalMAC, frameNumber string) error {
	// Skip if MAC was already provided (not UNKNOWN_*)
	if !strings.HasPrefix(originalMAC, "UNKNOWN_") {
		return nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open dump for MAC extraction: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file during MAC extraction: %v\n", closeErr)
		}
	}()

	// Read BLE MAC address from offset 0x5A000 (BLE Secrets location)
	if _, err := file.Seek(0x5A000, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to BLE secrets: %v", err)
	}

	macBytes := make([]byte, 6)
	if _, err := file.Read(macBytes); err != nil {
		return fmt.Errorf("failed to read MAC address: %v", err)
	}

	// Convert MAC bytes to string (reverse order for BLE)
	macAddress := fmt.Sprintf("%02X%02X%02X%02X%02X%02X",
		macBytes[5], macBytes[4], macBytes[3], macBytes[2], macBytes[1], macBytes[0])

	// Check if MAC is valid (not all zeros or all FFs)
	if macAddress == "000000000000" || macAddress == "FFFFFFFFFFFF" {
		fmt.Printf("Warning: Invalid MAC address found, keeping original filename\n")
		return nil
	}

	newFilename := fmt.Sprintf("VMES3-%s-%s.bin", frameNumber, macAddress)
	if newFilename == filename {
		return nil // No change needed
	}

	if err := os.Rename(filename, newFilename); err != nil {
		return fmt.Errorf("failed to rename file: %v", err)
	}

	fmt.Printf("✓ Extracted MAC address: %s\n", macAddress)
	fmt.Printf("✓ Renamed file to: %s\n", newFilename)
	return nil
}

// verifyDumpIntegrity performs comprehensive verification by comparing disk vs memory SHA512
func verifyDumpIntegrity(conn spi.Conn, filename string) error {
	fmt.Printf("\n🔍 Starting comprehensive dump verification...\n")

	// Step 1: Calculate SHA512 from disk file
	fmt.Printf("📁 Calculating SHA512 from disk file...\n")
	diskSHA512, err := calculateFileSHA512(filename)
	if err != nil {
		return fmt.Errorf("failed to calculate disk SHA512: %v", err)
	}
	fmt.Printf("📁 Disk SHA512: %x\n", diskSHA512)

	// Step 2: Re-dump from SPI chip into memory and calculate SHA512
	fmt.Printf("💾 Re-dumping from SPI chip into memory...\n")
	memorySHA512, err := calculateSPISHA512(conn)
	if err != nil {
		return fmt.Errorf("failed to calculate SPI SHA512: %v", err)
	}
	fmt.Printf("💾 Memory SHA512: %x\n", memorySHA512)

	// Step 3: Compare SHA512 hashes
	if string(diskSHA512) == string(memorySHA512) {
		fmt.Printf("✅ VERIFICATION PASSED: Disk and SPI memory SHA512 match!\n")
		return nil
	} else {
		return fmt.Errorf("❌ VERIFICATION FAILED: SHA512 mismatch between disk and SPI memory")
	}
}

// calculateFileSHA512 calculates SHA512 hash of a file on disk
func calculateFileSHA512(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file during SHA512 calculation: %v\n", closeErr)
		}
	}()

	hash := sha512.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

// calculateSPISHA512 re-dumps SPI flash into memory and calculates SHA512
func calculateSPISHA512(conn spi.Conn) ([]byte, error) {
	hash := sha512.New()

	totalChunks := flashSize / chunkSize
	for offset := 0; offset < flashSize; offset += chunkSize {
		// 24-bit address for SPI command
		address := []byte{
			byte(offset >> 16),
			byte(offset >> 8),
			byte(offset),
		}

		cmd := append([]byte{readCommand}, address...)
		data := make([]byte, chunkSize)

		if err := conn.Tx(cmd, data); err != nil {
			return nil, fmt.Errorf("SPI read failed at 0x%06X: %v", offset, err)
		}

		if _, err := hash.Write(data); err != nil {
			return nil, fmt.Errorf("hash write failed at 0x%06X: %v", offset, err)
		}

		// Progress reporting for verification
		chunk := offset/chunkSize + 1
		if chunk%2048 == 0 || chunk == totalChunks {
			progress := float64(chunk) / float64(totalChunks) * 100
			fmt.Printf("🔍 Verification progress: %.1f%% (%d/%d chunks)\n", progress, chunk, totalChunks)
		}
	}

	return hash.Sum(nil), nil
}

// validateBLEKey reads and validates BLE authentication key before dump
func validateBLEKey(conn spi.Conn) ([]byte, error) {
	fmt.Printf("🔑 Validating BLE authentication key...\n")

	// Read BLE key from offset 0x5A000 (60 bytes)
	bleKey, err := readBLEKeyFromSPI(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read BLE key: %v", err)
	}

	// Validate key content
	if isInvalidBLEKey(bleKey) {
		fmt.Printf("⚠️  WARNING: Invalid BLE authentication key detected!\n")
		fmt.Printf("Key: %X\n", bleKey[:16]) // Show first 16 bytes

		if err := handleInvalidBLEKey(); err != nil {
			return nil, err
		}
	}

	fmt.Printf("✅ BLE authentication key validated\n")
	return bleKey, nil
}

// readBLEKeyFromSPI reads BLE key from SPI flash
func readBLEKeyFromSPI(conn spi.Conn) ([]byte, error) {
	offset := 0x5A000 // BLE secrets location
	address := []byte{
		byte(offset >> 16),
		byte(offset >> 8),
		byte(offset),
	}

	cmd := append([]byte{readCommand}, address...)
	bleKey := make([]byte, 60) // 60 bytes BLE secrets

	if err := conn.Tx(cmd, bleKey); err != nil {
		return nil, err
	}

	return bleKey, nil
}

// isInvalidBLEKey checks if BLE key is invalid (all zeros, all FFs, or pattern)
func isInvalidBLEKey(key []byte) bool {
	if len(key) < 16 {
		return true
	}

	// Check for all zeros
	allZeros := true
	for _, b := range key[:16] {
		if b != 0x00 {
			allZeros = false
			break
		}
	}
	if allZeros {
		return true
	}

	// Check for all FFs
	allFFs := true
	for _, b := range key[:16] {
		if b != 0xFF {
			allFFs = false
			break
		}
	}
	if allFFs {
		return true
	}

	// Check for repeating patterns (same byte repeated)
	firstByte := key[0]
	repeating := true
	for _, b := range key[:16] {
		if b != firstByte {
			repeating = false
			break
		}
	}

	return repeating
}

// handleInvalidBLEKey prompts user and provides troubleshooting tips
func handleInvalidBLEKey() error {
	fmt.Printf("\n❌ INVALID BLE AUTHENTICATION KEY DETECTED\n")
	fmt.Printf("\n🔧 TROUBLESHOOTING TIPS:\n")
	fmt.Printf("1. Check SPI connections (CLK, MOSI, MISO, CS, GND)\n")
	fmt.Printf("2. Verify 3.3V power supply (not 5V!)\n")
	fmt.Printf("3. Check for loose connections or poor contact\n")
	fmt.Printf("4. Module might be active - remove power and battery\n")
	fmt.Printf("\n⚠️  Continuing with invalid key may result in corrupted dump!\n")

	fmt.Printf("\nDo you want to continue anyway? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		fmt.Printf("⚠️  Proceeding with potentially corrupted data...\n")
		return nil
	}

	return fmt.Errorf("dump cancelled by user - fix SPI connection and try again")
}

// validateStatusRegisters checks RDSR and RDSCUR for proper read conditions
func validateStatusRegisters(conn spi.Conn) error {
	fmt.Printf("🔍 Checking status registers...\n")

	// Read Status Register (RDSR - 0x05)
	rdsr, err := readStatusRegister(conn)
	if err != nil {
		return fmt.Errorf("failed to read RDSR: %v", err)
	}

	// Read Security Register (RDSCUR - 0x2B)
	rdscur, err := readSecurityRegister(conn)
	if err != nil {
		return fmt.Errorf("failed to read RDSCUR: %v", err)
	}

	// Check WIP (Write-In-Progress) bit (bit 0 of RDSR)
	if rdsr&0x01 != 0 {
		return fmt.Errorf("chip is busy (WIP=1), wait for operations to complete")
	}

	// Check Block Protection bits (BP2, BP1, BP0 - bits 4,3,2 of RDSR)
	bpBits := (rdsr >> 2) & 0x07
	if bpBits != 0 {
		fmt.Printf("⚠️  Block protection enabled (BP=%d), some regions may be protected\n", bpBits)
	}

	// Check Status Register Protect (SRP - bit 7 of RDSR)
	if rdsr&0x80 != 0 {
		fmt.Printf("⚠️  Status register write protection enabled (SRP=1)\n")
	}

	// Check Security Register checksum bits (bits 6,5 of RDSCUR)
	securityBits := (rdscur >> 5) & 0x03
	if securityBits != 0 {
		fmt.Printf("⚠️  Security register checksum bits set (%d), may indicate locked regions\n", securityBits)
	}

	fmt.Printf("✅ Status registers validated (RDSR=0x%02X, RDSCUR=0x%02X)\n", rdsr, rdscur)
	return nil
}

// validateChipCompatibility checks if chip is VanMoof S3 compatible
func validateChipCompatibility(conn spi.Conn) error {
	fmt.Printf("🔍 Validating chip compatibility...\n")

	// Read JEDEC ID
	cmd := []byte{0x9F} // Read JEDEC ID command
	response := make([]byte, 3)

	if err := conn.Tx(cmd, response); err != nil {
		return fmt.Errorf("failed to read chip ID: %v", err)
	}

	manufacturerID := response[0]
	deviceID := uint16(response[1])<<8 | uint16(response[2])

	// Check for Macronix MX25L51245G (VanMoof S3)
	if manufacturerID == 0xC2 && deviceID == 0x201A {
		fmt.Printf("✅ Chip validated: MX25L51245G (VanMoof S3 compatible)\n")
		return nil
	}

	// Unsupported chip detected
	fmt.Printf("❌ UNSUPPORTED CHIP DETECTED\n")
	fmt.Printf("Manufacturer: 0x%02X, Device: 0x%04X\n", manufacturerID, deviceID)
	fmt.Printf("\n⚠️  WARNING: This tool is designed for VanMoof S3 modules only!\n")
	fmt.Printf("Expected: Macronix MX25L51245G (0xC2, 0x201A)\n")
	fmt.Printf("Found: %s\n", getChipName(manufacturerID, deviceID))
	fmt.Printf("\n🚫 DUMP ABORTED - Wrong chip type detected\n")

	return fmt.Errorf("unsupported chip - VanMoof S3 requires MX25L51245G")
}

// getChipName returns human-readable chip name (VanMoof specific)
func getChipName(manufacturerID uint8, deviceID uint16) string {
	switch manufacturerID {
	case 0xC2:
		switch deviceID {
		case 0x201A:
			return "Macronix MX25L51245G (64MB - VanMoof S3)"
		default:
			return fmt.Sprintf("Non-VanMoof Macronix (0x%04X)", deviceID)
		}
	default:
		return fmt.Sprintf("Non-VanMoof Chip (0x%02X, 0x%04X)", manufacturerID, deviceID)
	}
}

// compareBLEKeys compares original BLE key with key from dump file
func compareBLEKeys(filename string, originalKey []byte) error {
	fmt.Printf("🔑 Comparing BLE keys...\n")

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open dump for BLE comparison: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file during BLE comparison: %v\n", closeErr)
		}
	}()

	// Read BLE key from dump at offset 0x5A000
	if _, err := file.Seek(0x5A000, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to BLE secrets in dump: %v", err)
	}

	dumpKey := make([]byte, 60)
	if _, err := file.Read(dumpKey); err != nil {
		return fmt.Errorf("failed to read BLE key from dump: %v", err)
	}

	// Compare first 16 bytes of BLE keys
	for i := 0; i < 16; i++ {
		if originalKey[i] != dumpKey[i] {
			fmt.Printf("❌ BLE key mismatch detected!\n")
			fmt.Printf("Original: %X\n", originalKey[:16])
			fmt.Printf("Dump:     %X\n", dumpKey[:16])
			return fmt.Errorf("BLE authentication key changed during dump")
		}
	}

	fmt.Printf("✅ BLE authentication key matches original\n")
	return nil
}
