package vanmoof

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DumpKeysAndChecksums extracts keys and generates SHA512 checksums after SPI flash dump
func DumpKeysAndChecksums(dumpFilename string) error {
	file, err := os.Open(dumpFilename)
	if err != nil {
		return fmt.Errorf("failed to open dump file: %v", err)
	}
	defer file.Close()

	baseName := strings.TrimSuffix(dumpFilename, filepath.Ext(dumpFilename))

	// Extract and save keys
	if err := extractKeysToFile(file, baseName); err != nil {
		return fmt.Errorf("failed to extract keys: %v", err)
	}

	// Generate SHA512 checksum
	if err := generateSHA512File(dumpFilename, baseName); err != nil {
		return fmt.Errorf("failed to generate SHA512: %v", err)
	}

	fmt.Printf("✅ Keys and checksums saved successfully\n")
	return nil
}

// extractKeysToFile extracts all keys from dump and saves to .keys file
func extractKeysToFile(file *os.File, baseName string) error {
	keysFilename := baseName + ".keys"
	keysFile, err := os.Create(keysFilename)
	if err != nil {
		return fmt.Errorf("failed to create keys file: %v", err)
	}
	defer keysFile.Close()

	fmt.Fprintf(keysFile, "# VanMoof Module Keys - Extracted from SPI Flash Dump\n")
	fmt.Fprintf(keysFile, "# Generated: %s\n\n", getCurrentTimestamp())

	// BLE Authentication Key (16 bytes at 0x005A000)
	bleKey := readFromFile(file, 0x005A000, 16)
	fmt.Fprintf(keysFile, "BLE_AUTH_KEY=%s\n", strings.ToUpper(hex.EncodeToString(bleKey)))

	// Manufacturing Key (16 bytes at 0x005AFC0)
	mfgKey := readFromFile(file, 0x005AFC0, 16)
	fmt.Fprintf(keysFile, "MFG_KEY=%s\n", strings.ToUpper(hex.EncodeToString(mfgKey)))

	// M-ID/M-KEY (60 bytes at 0x005af80)
	midKey := readFromFile(file, 0x005af80, 60)
	fmt.Fprintf(keysFile, "M_ID_KEY=%s\n", hex.EncodeToString(midKey))

	// MAC Address with MOOF validation
	macBuf := readFromFile(file, 0x0005AFE0, 16)
	macStr := string(macBuf)
	if strings.HasSuffix(macStr, "MOOF") {
		mac := macStr[:12]
		fmt.Fprintf(keysFile, "MAC_ADDRESS=%s\n", mac)
	} else {
		fmt.Fprintf(keysFile, "MAC_ADDRESS=INVALID_NO_MOOF_SIGNATURE\n")
	}

	fmt.Printf("🔑 Keys extracted to: %s\n", keysFilename)
	return nil
}

// generateSHA512File calculates SHA512 of dump file and saves to .sha512 file
func generateSHA512File(dumpFilename, baseName string) error {
	file, err := os.Open(dumpFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha512.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	sha512sum := hex.EncodeToString(hash.Sum(nil))

	checksumFilename := baseName + ".sha512"
	checksumFile, err := os.Create(checksumFilename)
	if err != nil {
		return fmt.Errorf("failed to create checksum file: %v", err)
	}
	defer checksumFile.Close()

	fmt.Fprintf(checksumFile, "%s  %s\n", sha512sum, filepath.Base(dumpFilename))

	fmt.Printf("🔐 SHA512 checksum saved to: %s\n", checksumFilename)
	fmt.Printf("🔐 SHA512: %s\n", sha512sum)
	return nil
}

// getCurrentTimestamp returns current timestamp in YYYYMMDD-HHMMSS format
func getCurrentTimestamp() string {
	return time.Now().Format("20060102-150405")
}
