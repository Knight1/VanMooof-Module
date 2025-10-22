package vanmoof

import (
	"flag"
	"fmt"
	"time"

	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

var debugLogging = flag.Bool("d", false, "Enable debug logging")

// sendCommand sends a single SPI command with optional data.
func sendCommand(conn spi.Conn, cmd byte, data []byte) error {
	payload := append([]byte{cmd}, data...)
	if err := conn.Tx(payload, nil); err != nil {
		return fmt.Errorf("SPI command failed: %v", err)
	}
	return nil
}

// waitForWriteComplete polls the status register until the write is done.
func waitForWriteComplete(conn spi.Conn) error {
	for {
		// Read Status Register command: 0x05
		status := make([]byte, 1)
		if err := conn.Tx([]byte{0x05}, status); err != nil {
			return fmt.Errorf("failed to read status register: %v", err)
		}

		// Check the Write-In-Progress (WIP) bit (bit 0 of status register).
		if status[0]&0x01 == 0 {
			return nil // Write complete
		}

		// Wait a bit before polling again.
		time.Sleep(10 * time.Millisecond)
	}
}

func spiConnect() (conn spi.Conn, err error) {
	if _, err := host.Init(); err != nil {
		fmt.Printf("Failed to initialize host: %v\n", err)
		return nil, err
	}

	p, err := spireg.Open("")
	if err != nil {
		fmt.Printf("Failed to open SPI port: %v\n", err)
		return nil, err
	}
	defer p.Close()

	conn, err = p.Connect(10000000, spi.Mode0, 8)
	if err != nil {
		fmt.Printf("Failed to configure SPI connection: %v\n", err)
		return nil, err
	}

	if *debugLogging {
		fmt.Println("Got SPI Flash Chip Connection")
		fmt.Println("Connection:", conn)
	}

	return conn, nil
}

// writeEnable sends the Write Enable command (0x06) to the SPI flash.
func writeEnable(conn spi.Conn) error {
	if err := sendCommand(conn, 0x06, nil); err != nil {
		return fmt.Errorf("Failed to send Write Enable command: %v", err)
	}

	if *debugLogging {
		fmt.Println("Write Enable Command successfully sent to the SPI Flash")
	}

	enabled, err := verifyWriteEnable(conn)
	if err != nil {
		fmt.Println("Failed to Enable 'Write Enable':", err)
	}

	if !enabled {
		fmt.Println("Write Enable was successfully but chip is still write Disabled.")
		return nil
	}
	return nil
}

// writeDisable sends the Write Disable command (0x04) to the SPI flash.
func writeDisable(conn spi.Conn) error {
	if err := sendCommand(conn, 0x04, nil); err != nil {
		return fmt.Errorf("Failed to send Write Disable command: %v", err)
	}

	if *debugLogging {
		fmt.Println("Write Disable Command successfully sent to the SPI Flash")
	}

	enabled, err := verifyWriteEnable(conn)
	if err != nil {
		fmt.Println("Failed to Disable 'Write Enable':", err)
	}

	if enabled {
		fmt.Println("Write Disable was successfully but chip is still write Enabled.")
		return nil
	}
	return nil
}

// verifyWriteEnable checks if the Write Enable Latch (WEL) bit is set.
func verifyWriteEnable(conn spi.Conn) (bool, error) {
	status := make([]byte, 1)

	// Send the Read Status Register command (0x05).
	if err := conn.Tx([]byte{0x05}, status); err != nil {
		return false, fmt.Errorf("failed to read status register: %v", err)
	}

	// Check if the WEL bit (bit 1) is set.
	return (status[0] & 0x02) != 0, nil
}
