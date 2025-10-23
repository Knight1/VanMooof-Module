package vanmoof

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/notifai/ymodem/ymodem"
)

const (
	packMagic   = "PACK"
	maxPackSize = 2 * 1024 * 1024 // 2MB max PACK size
)

// SerialPort interface for cross-platform serial communication
type SerialPort interface {
	io.ReadWriter
	Close() error
}

// ValidatePACK validates a PACK file and ensures it's not an SPI dump
func ValidatePACK(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file: %v\n", closeErr)
		}
	}()

	// Check file size - SPI dumps are typically 64MB
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	if stat.Size() > maxPackSize {
		return fmt.Errorf("file too large (%d bytes), PACK files must be ≤2MB", stat.Size())
	}

	// Read and validate PACK header
	header := make([]byte, 12)
	_, err = io.ReadFull(file, header)
	if err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}

	// Check magic bytes
	if !bytes.Equal(header[0:4], []byte(packMagic)) {
		return fmt.Errorf("invalid PACK magic, expected %q got %q", packMagic, string(header[0:4]))
	}

	// Parse header
	offset := binary.LittleEndian.Uint32(header[4:8])
	length := binary.LittleEndian.Uint32(header[8:12])

	// Validate header values
	if offset == 0 || length == 0 {
		return fmt.Errorf("invalid PACK header: offset=%d, length=%d", offset, length)
	}

	if int64(offset)+int64(length)+12 > stat.Size() {
		return fmt.Errorf("PACK directory beyond file bounds: offset=%d + length=%d + 12 > filesize=%d", offset, length, stat.Size())
	}

	// Validate directory structure
	_, err = file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to directory: %v", err)
	}

	entrySize := 64 // 56 bytes filename + 4 bytes offset + 4 bytes length
	entryCount := int(length) / entrySize

	if entryCount == 0 {
		return fmt.Errorf("no entries found in PACK directory")
	}

	// Validate at least one entry
	entry := make([]byte, entrySize)
	_, err = io.ReadFull(file, entry)
	if err != nil {
		return fmt.Errorf("failed to read directory entry: %v", err)
	}

	// Check if filename is valid (not all zeros) - fix type conversion
	filenameBytes := bytes.TrimRight(entry[0:56], "\x00")
	if len(filenameBytes) == 0 {
		return fmt.Errorf("invalid PACK entry: empty filename")
	}

	fmt.Printf("PACK validation successful: %d entries found\n", entryCount)
	return nil
}

// UploadPACK uploads a PACK file via Y-Modem to the specified serial port
func UploadPACK(packFile, serialPort string, baudRate uint32) error {
	// Validate PACK file first
	err := ValidatePACK(packFile)
	if err != nil {
		return fmt.Errorf("PACK validation failed: %v", err)
	}

	// Read PACK file
	packData, err := os.ReadFile(packFile)
	if err != nil {
		return fmt.Errorf("failed to read PACK file: %v", err)
	}

	// Open serial port (platform-specific)
	port, err := openSerial(serialPort, baudRate)
	if err != nil {
		return fmt.Errorf("failed to open serial port: %v", err)
	}
	defer func() {
		if closeErr := port.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close serial port: %v\n", closeErr)
		}
	}()

	fmt.Printf("Uploading %s (%d bytes) to %s at 115200 baud...\n",
		filepath.Base(packFile), len(packData), serialPort)

	// Prepare Y-Modem file
	files := []ymodem.File{
		{
			Name: "pack.bin",
			Data: packData,
		},
	}

	// Send via Y-Modem with 1024 byte blocks
	err = ymodem.ModemSend(port, nil, 1024, files)
	if err != nil {
		return fmt.Errorf("Y-Modem transfer failed: %v", err)
	}

	fmt.Println("PACK upload completed successfully!")
	return nil
}

// ListSerialPorts lists available serial ports (platform-specific)
func ListSerialPorts() ([]string, error) {
	switch runtime.GOOS {
	case "windows":
		return listWindowsPorts()
	case "darwin", "linux":
		return listUnixPorts()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// listUnixPorts lists serial ports on Unix-like systems (macOS/Linux)
func listUnixPorts() ([]string, error) {
	var ports []string
	var patterns []string

	switch runtime.GOOS {
	case "darwin":
		// macOS patterns
		patterns = []string{
			"/dev/tty.usbserial*",
			"/dev/tty.usbmodem*",
			"/dev/tty.SLAB_USBtoUART*",
			"/dev/tty.wchusbserial*",
		}
	case "linux":
		// Linux patterns
		patterns = []string{
			"/dev/ttyUSB*",
			"/dev/ttyACM*",
			"/dev/ttyS*",
		}
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("Warning: failed to glob pattern %s: %v\n", pattern, err)
			continue
		}
		ports = append(ports, matches...)
	}

	return ports, nil
}

// listWindowsPorts lists COM ports on Windows
func listWindowsPorts() ([]string, error) {
	var ports []string

	// Check COM1 through COM20
	for i := 1; i <= 20; i++ {
		portName := fmt.Sprintf("COM%d", i)
		if port, err := openSerial(portName, 115200); err == nil {
			port.Close()
			ports = append(ports, portName)
		}
	}

	return ports, nil
}

// GetDefaultSerialPort returns the default serial port for the platform
func GetDefaultSerialPort() string {
	switch runtime.GOOS {
	case "windows":
		return "COM3"
	case "darwin":
		return "/dev/tty.usbserial-0001"
	case "linux":
		return "/dev/ttyUSB0"
	default:
		return "COM3"
	}
}

// ValidateSerialPort checks if a serial port name is valid for the platform
func ValidateSerialPort(port string) error {
	switch runtime.GOOS {
	case "windows":
		if !strings.HasPrefix(strings.ToUpper(port), "COM") {
			return fmt.Errorf("Windows serial ports must start with COM (e.g., COM3)")
		}
	case "darwin":
		if !strings.HasPrefix(port, "/dev/tty.") {
			return fmt.Errorf("macOS serial ports must start with /dev/tty. (e.g., /dev/tty.usbserial-0001)")
		}
	case "linux":
		if !strings.HasPrefix(port, "/dev/") {
			return fmt.Errorf("Linux serial ports must start with /dev/ (e.g., /dev/ttyUSB0)")
		}
	}
	return nil
}
