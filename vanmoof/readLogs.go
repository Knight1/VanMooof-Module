package vanmoof

import (
	"fmt"
	"os"
)

func readFromFile(file *os.File, offset int64, length int) []byte {
	buf := make([]byte, length)
	_, err := file.ReadAt(buf, offset)
	if err != nil {
		fmt.Printf("Error reading from file: %v\n", err)
		return nil
	}
	return buf
}

func ReadLogs(file *os.File) {
	offset := int64(0x3fdd000)

	// Get file size to avoid EOF
	stat, err := file.Stat()
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return
	}

	fileSize := stat.Size()
	if offset >= fileSize {
		fmt.Printf("Log offset 0x%x is beyond file size 0x%x\n", offset, fileSize)
		return
	}

	// Calculate available bytes from offset to end of file
	availableBytes := fileSize - offset
	length := int(availableBytes)
	if length > 1024*1024 {
		length = 1024 * 1024 // Cap at 1MB
	}

	fmt.Printf("Reading %d bytes from offset 0x%x (file size: 0x%x)\n", length, offset, fileSize)

	buf := readFromFile(file, offset, length)
	if buf == nil {
		return
	}

	// Parse ASCII log entries
	var logEntries []string
	start := 0

	for i := 0; i < len(buf)-10; i++ {
		// Look for FFFFFFFFFFF... pattern (end of log entry)
		if buf[i] == 0xFF && buf[i+1] == 0xFF && buf[i+2] == 0xFF {
			// Found end marker, extract log entry
			if start < i {
				entry := string(buf[start:i])
				// Only add non-empty entries
				if len(entry) > 0 && entry[0] != 0xFF {
					logEntries = append(logEntries, entry)
				}
			}
			// Skip past the FFFFFF... marker
			for i < len(buf) && buf[i] == 0xFF {
				i++
			}
			start = i
		}
	}

	// Print log entries
	fmt.Printf("Found %d log entries:\n", len(logEntries))
	for i, entry := range logEntries {
		fmt.Printf("Log %d: %s\n", i+1, entry)
	}
}
