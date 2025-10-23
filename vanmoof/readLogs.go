package vanmoof

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

func readFromFile(file *os.File, offset int64, length int) []byte {
	buf := make([]byte, length)
	n, err := file.ReadAt(buf, offset)
	if err != nil && err.Error() != "EOF" {
		fmt.Printf("Error reading from file: %v\n", err)
		return nil
	}
	if n < length {
		// Return only the bytes that were actually read
		return buf[:n]
	}
	return buf
}

// isPrintableASCII checks if a string contains mostly printable ASCII characters
func isPrintableASCII(s string) bool {
	if len(s) == 0 {
		return false
	}
	printableCount := 0
	for _, r := range s {
		if unicode.IsPrint(r) || r == '\n' || r == '\r' || r == '\t' {
			printableCount++
		}
	}
	// Require at least 80% printable characters
	return float64(printableCount)/float64(len(s)) >= 0.8
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

	for i := 0; i < len(buf)-2; i++ {
		// Look for null terminator or FFFFFF pattern (end of log entry)
		if buf[i] == 0x00 || (buf[i] == 0xFF && buf[i+1] == 0xFF) {
			// Found end marker, extract log entry
			if start < i {
				entry := string(buf[start:i])
				// Clean up the entry and only add non-empty, printable entries
				entry = strings.TrimSpace(entry)
				if len(entry) > 0 && isPrintableASCII(entry) {
					logEntries = append(logEntries, entry)
				}
			}
			// Skip past the marker
			for i < len(buf) && (buf[i] == 0x00 || buf[i] == 0xFF) {
				i++
			}
			start = i
		}
	}

	// Check for any remaining entry at the end
	if start < len(buf) {
		entry := string(buf[start:])
		entry = strings.TrimSpace(entry)
		if len(entry) > 0 && isPrintableASCII(entry) {
			logEntries = append(logEntries, entry)
		}
	}

	// Print log entries
	fmt.Printf("Found %d log entries:\n", len(logEntries))
	if len(logEntries) > 0 {
		for i, entry := range logEntries {
			fmt.Printf("Log %d: \n%s\n", i+1, entry)
		}
	}
}

// ReadLogsCount only shows the count of log entries for the -show command
func ReadLogsCount(file *os.File) {
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

	// Parse ASCII log entries (same logic as ReadLogs but only count)
	var logCount int
	start := 0

	for i := 0; i < len(buf)-2; i++ {
		// Look for null terminator or FFFFFF pattern (end of log entry)
		if buf[i] == 0x00 || (buf[i] == 0xFF && buf[i+1] == 0xFF) {
			// Found end marker, extract log entry
			if start < i {
				entry := string(buf[start:i])
				// Clean up the entry and only count non-empty, printable entries
				entry = strings.TrimSpace(entry)
				if len(entry) > 0 && isPrintableASCII(entry) {
					logCount++
				}
			}
			// Skip past the marker
			for i < len(buf) && (buf[i] == 0x00 || buf[i] == 0xFF) {
				i++
			}
			start = i
		}
	}

	// Check for any remaining entry at the end
	if start < len(buf) {
		entry := string(buf[start:])
		entry = strings.TrimSpace(entry)
		if len(entry) > 0 && isPrintableASCII(entry) {
			logCount++
		}
	}

	// Only print the count
	fmt.Printf("Found %d log entries:\n", logCount)
}
