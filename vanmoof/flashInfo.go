package vanmoof

import (
	"fmt"
	"periph.io/x/conn/v3/spi"
)

const (
	// SPI Flash commands
	cmdReadID  = 0x9F // Read JEDEC ID
	cmdReadUID = 0x4B // Read Unique ID (if supported)
	cmdReadSN  = 0xC0 // Read Serial Number (Macronix specific)
)

// FlashInfo contains SPI flash chip information
type FlashInfo struct {
	ManufacturerID uint8
	DeviceID       uint16
	Manufacturer   string
	DeviceName     string
	Capacity       string
	SerialNumber   []byte
	UniqueID       []byte
}

// ReadFlashInfo reads the SPI flash chip identification and serial number
func ReadFlashInfo(sudo bool) (*FlashInfo, error) {
	if !sudo {
		return nil, fmt.Errorf("SPI flash access requires -sudo flag")
	}

	conn, err := spiConnect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SPI: %v", err)
	}

	// Check status registers for proper read conditions
	if err := validateStatusRegisters(conn); err != nil {
		return nil, fmt.Errorf("status register validation failed: %v", err)
	}

	info := &FlashInfo{}

	// Read JEDEC ID (Manufacturer + Device ID)
	if err := readJEDECID(conn, info); err != nil {
		return nil, fmt.Errorf("failed to read JEDEC ID: %v", err)
	}

	// Try to read serial number (Macronix specific)
	if info.ManufacturerID == 0xC2 { // Macronix
		if err := readMacronixSerialNumber(conn, info); err != nil {
			fmt.Printf("Warning: Could not read serial number: %v\n", err)
		}
	}

	// Try to read unique ID
	if err := readUniqueID(conn, info); err != nil {
		fmt.Printf("Warning: Could not read unique ID: %v\n", err)
	}

	return info, nil
}

// readJEDECID reads the JEDEC manufacturer and device ID
func readJEDECID(conn spi.Conn, info *FlashInfo) error {
	cmd := []byte{cmdReadID}
	response := make([]byte, 3)

	if err := conn.Tx(cmd, response); err != nil {
		return err
	}

	info.ManufacturerID = response[0]
	info.DeviceID = uint16(response[1])<<8 | uint16(response[2])

	// Decode manufacturer (VanMoof only uses Macronix)
	switch info.ManufacturerID {
	case 0xC2:
		info.Manufacturer = "Macronix"
	default:
		info.Manufacturer = fmt.Sprintf("Non-VanMoof (0x%02X)", info.ManufacturerID)
	}

	// Decode Macronix device (VanMoof uses 64MB MX25L51245G)
	if info.ManufacturerID == 0xC2 {
		switch info.DeviceID {
		case 0x201A:
			info.DeviceName = "MX25L51245G"
			info.Capacity = "64MB (512Mbit)"
		default:
			info.DeviceName = fmt.Sprintf("Non-VanMoof Macronix (0x%04X)", info.DeviceID)
		}
	} else {
		info.DeviceName = fmt.Sprintf("Device ID: 0x%04X", info.DeviceID)
	}

	return nil
}

// readMacronixSerialNumber reads Macronix-specific serial number
func readMacronixSerialNumber(conn spi.Conn, info *FlashInfo) error {
	// Macronix serial number command (if supported)
	cmd := []byte{cmdReadSN, 0x00, 0x00, 0x00, 0x00} // Command + 4 dummy bytes
	response := make([]byte, 8)                      // 8-byte serial number

	if err := conn.Tx(cmd, response); err != nil {
		return err
	}

	// Check if we got valid data (not all 0xFF or 0x00)
	allFF := true
	allZero := true
	for _, b := range response {
		if b != 0xFF {
			allFF = false
		}
		if b != 0x00 {
			allZero = false
		}
	}

	if !allFF && !allZero {
		info.SerialNumber = response
	}

	return nil
}

// readUniqueID reads the unique ID if supported
func readUniqueID(conn spi.Conn, info *FlashInfo) error {
	cmd := []byte{cmdReadUID, 0x00, 0x00, 0x00, 0x00} // Command + 4 dummy bytes
	response := make([]byte, 8)                       // 8-byte unique ID

	if err := conn.Tx(cmd, response); err != nil {
		return err
	}

	// Check if we got valid data
	allFF := true
	allZero := true
	for _, b := range response {
		if b != 0xFF {
			allFF = false
		}
		if b != 0x00 {
			allZero = false
		}
	}

	if !allFF && !allZero {
		info.UniqueID = response
	}

	return nil
}

// String returns a formatted string representation of flash info
func (info *FlashInfo) String() string {
	result := fmt.Sprintf("SPI Flash Information:\n")
	result += fmt.Sprintf("  Manufacturer: %s (0x%02X)\n", info.Manufacturer, info.ManufacturerID)
	result += fmt.Sprintf("  Device: %s\n", info.DeviceName)
	result += fmt.Sprintf("  Capacity: %s\n", info.Capacity)

	if len(info.SerialNumber) > 0 {
		result += fmt.Sprintf("  Serial Number: %X\n", info.SerialNumber)
	}

	if len(info.UniqueID) > 0 {
		result += fmt.Sprintf("  Unique ID: %X\n", info.UniqueID)
	}

	return result
}
