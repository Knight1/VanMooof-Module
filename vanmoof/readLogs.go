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

	availableBytes := fileSize - offset
	length := int(availableBytes)
	if length > 1024*1024 {
		length = 1024 * 1024
	}

	fmt.Printf("Reading %d bytes from offset 0x%x (file size: 0x%x)\n", length, offset, fileSize)

	buf := readFromFile(file, offset, length)
	if buf == nil {
		return
	}

	blocks := findLogBlocks(buf)
	totalLines := 0
	for _, b := range blocks {
		totalLines += countLogLines(buf[b[0]:b[1]])
	}

	fmt.Printf("Found %d log blocks containing %d log entries\n", len(blocks), totalLines)
	for i, b := range blocks {
		blockBytes := buf[b[0]:b[1]]
		lines := countLogLines(blockBytes)
		fmt.Printf("\n--- Block %d @ 0x%08X (%d bytes, %d entries) ---\n",
			i+1, offset+int64(b[0]), b[1]-b[0], lines)
		// Print the block as text, with NULs treated as line breaks.
		text := strings.ReplaceAll(string(blockBytes), "\x00", "\n")
		fmt.Println(strings.TrimRight(text, "\n"))
	}
}

// logFFGapThreshold is the minimum run of consecutive 0xFF bytes
// that the log writer leaves between log "blocks" (one per boot
// cycle / log flush). Anything shorter is treated as content
// (could be a binary field inside a log line). Tuned empirically:
// real entries never carry 4 consecutive 0xFF in a row, while
// inter-block padding is hundreds-to-thousands of bytes long.
const logFFGapThreshold = 4

// findLogBlocks returns the [start, end) byte offsets (relative
// to buf) of each non-FF "log block" — i.e. each region of the log
// buffer separated from its neighbours by a run of at least
// logFFGapThreshold consecutive 0xFF bytes. Blocks that are
// entirely 0x00 padding are dropped (uninitialised flash that
// was zeroed rather than erased).
func findLogBlocks(buf []byte) [][2]int {
	var blocks [][2]int
	i, n := 0, len(buf)
	for i < n {
		// Skip leading FF padding.
		for i < n && buf[i] == 0xFF {
			i++
		}
		if i >= n {
			break
		}
		start := i
		// Advance until we hit a long-enough FF run.
		for i < n {
			if buf[i] != 0xFF {
				i++
				continue
			}
			runStart := i
			for i < n && buf[i] == 0xFF {
				i++
			}
			if i-runStart >= logFFGapThreshold {
				blocks = append(blocks, [2]int{start, runStart})
				start = -1
				break
			}
		}
		if start >= 0 {
			blocks = append(blocks, [2]int{start, i})
		}
	}

	// Filter out blocks that are entirely 0x00 — those are pre-
	// initialised flash, not real log data.
	out := blocks[:0]
	for _, b := range blocks {
		nonZero := false
		for _, c := range buf[b[0]:b[1]] {
			if c != 0x00 {
				nonZero = true
				break
			}
		}
		if nonZero {
			out = append(out, b)
		}
	}
	return out
}

// countLogLines counts newline- or null-terminated entries inside
// a single log block, ignoring lines that aren't substantially
// printable ASCII. A trailing partial line (no terminator before
// the FF gap) is counted as one entry.
func countLogLines(block []byte) int {
	count := 0
	start := 0
	for i := 0; i <= len(block); i++ {
		end := i == len(block) || block[i] == '\n' || block[i] == '\r' || block[i] == 0x00
		if !end {
			continue
		}
		if start < i {
			line := strings.TrimSpace(string(block[start:i]))
			if line != "" && isPrintableASCII(line) {
				count++
			}
		}
		start = i + 1
	}
	return count
}

// ReadLogsCount summarises the log region: how many distinct log
// blocks (boot-cycle dumps) were written and how many individual
// log lines they contain in total. Blocks are recognised by the
// long 0xFF gaps the firmware leaves between flushes.
func ReadLogsCount(file *os.File) {
	offset := int64(0x3fdd000)

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

	availableBytes := fileSize - offset
	length := int(availableBytes)
	if length > 1024*1024 {
		length = 1024 * 1024
	}

	fmt.Printf("Reading %d bytes from offset 0x%x (file size: 0x%x)\n", length, offset, fileSize)

	buf := readFromFile(file, offset, length)
	if buf == nil {
		return
	}

	blocks := findLogBlocks(buf)
	totalLines := 0
	for _, b := range blocks {
		totalLines += countLogLines(buf[b[0]:b[1]])
	}

	fmt.Printf("Found %d log blocks containing %d log entries\n", len(blocks), totalLines)
	if len(blocks) == 0 {
		return
	}
	for i, b := range blocks {
		lines := countLogLines(buf[b[0]:b[1]])
		fmt.Printf("  block %d: 0x%08X - 0x%08X (%d bytes, %d entries)\n",
			i+1, offset+int64(b[0]), offset+int64(b[1]), b[1]-b[0], lines)
	}
}
