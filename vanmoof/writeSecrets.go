package vanmoof

import (
	"encoding/hex"
	"flag"
	"fmt"
)

var sudo = flag.Bool("iKnowWhatIAmDoingISwear", false, "Use sudo")

// VERY DANGEROUS!
func WriteSecrets(secret string, value string) {
	// Check if the secret is unlock
	if secret != "unlock" {
		fmt.Println("unknown write Operation")
		return
	}

	fmt.Println("Begin to write", value, "into", secret, "Chip Location")

	if !*sudo {
		fmt.Println("Using the Write Secrets Function without a VALID and KNOWN good Backup of the SPI Flash is EXTREMELY DANGEROUS!")
		fmt.Println("Supply the sudo Mode to continue.")
		return
	}

	// check the value for 16 bytes of hex
	if len(value) != 16 {
		fmt.Println("value must be 16 bytes without spaces.")
		return
	}

	valueHex, err := hex.DecodeString(value)
	if err != nil {
		fmt.Println("Failed to decode String to Hexadecimal:", err)
	}

	conn, err := spiConnect()
	if err != nil {
		fmt.Println("")
	}

	// Enable writing to flash
	if err := writeEnable(conn); err != nil {
		fmt.Printf("Failed to enable write: %v\n", err)
		return
	}

	address := []byte{0x00, 0x5A, 0x00} // 24-bit address 0x005A000

	// Write 16 bytes to address 0x005A000
	if err := sendCommand(conn, 0x02, append(address, valueHex...)); err != nil {
		fmt.Printf("Failed to send Page Program command: %v\n", err)
		return
	}

	// Wait for write operation to complete by polling the status register.
	if err := waitForWriteComplete(conn); err != nil {
		fmt.Printf("Write operation failed: %v\n", err)
		return
	}

	// After writing, disable write access for protection.
	if err := writeDisable(conn); err != nil {
		fmt.Printf("Failed to disable write: %v\n", err)
		return
	}

	fmt.Println("Data written successfully!")

}
