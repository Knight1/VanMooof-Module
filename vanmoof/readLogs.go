package vanmoof

import (
	"fmt"
	"os"
	"strconv"
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
	offset := int64(logRegionBase)

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

	length := logRegionSize
	if int(fileSize-offset) < length {
		length = int(fileSize - offset)
	}

	fmt.Printf("Reading %d bytes from log region @ 0x%X (128 KB circular buffer, 16-byte entries)\n",
		length, offset)

	buf := readFromFile(file, offset, length)
	if buf == nil {
		return
	}

	blocks := findLogBlocks(buf)
	totalLines := 0
	for _, b := range blocks {
		totalLines += countLogLines(buf[b[0]:b[1]])
	}

	fmt.Printf("Log region: %d log blocks containing %d \\n-terminated entries\n",
		len(blocks), totalLines)
	for i, b := range blocks {
		blockBytes := buf[b[0]:b[1]]
		lines := countLogLines(blockBytes)
		fmt.Printf("\n--- log [%d] @ 0x%08X (%d bytes, %d entries) ---\n",
			i, offset+int64(b[0]), b[1]-b[0], lines)
		text := strings.ReplaceAll(string(blockBytes), "\x00", "\n")
		fmt.Println(strings.TrimRight(text, "\n"))
	}
}

// Log region layout per bleware/src/log_gatt.c (DAT_0001EA34 +
// LOG_REGION_BASE/SIZE) and bleware/src/monitor/cmd_log.c:
//
//   0x03FDC000 - 0x03FDD000   4 KB   cursor persistence sector
//                                     (circular array of (head, tail)
//                                      u32 pairs, 8 bytes per write,
//                                      wraps mod 0x1000)
//   0x03FDD000 - 0x03FFD000  128 KB  log circular buffer
//                                     (16-byte entries; lines are
//                                      \n-terminated ASCII spanning
//                                      multiple entries)
//   0x03FFD000 - 0x04000000   12 KB  unaccounted tail
//
// Wrap mask for the log region is 0x1FFFF; log_read_entry indexes
// into it in 16-byte units.
const (
	logRegionBase   = 0x03FDD000
	logRegionSize   = 0x00020000
	logEntrySize    = 16
	logCursorSector = 0x03FDC000
	logCursorSize   = 0x00001000
)

// countLogLines counts \n-terminated entries inside the populated
// portion of the log buffer. \r and NUL are also treated as line
// terminators (the firmware emits CRLF on the UART path but plain
// LF on the BLE-readout path). Lines that aren't substantially
// printable ASCII are skipped.
func countLogLines(buf []byte) int {
	count := 0
	start := 0
	for i := 0; i <= len(buf); i++ {
		end := i == len(buf) || buf[i] == '\n' || buf[i] == '\r' || buf[i] == 0x00
		if !end {
			continue
		}
		if start < i {
			line := strings.TrimSpace(string(buf[start:i]))
			if line != "" && isPrintableASCII(line) {
				count++
			}
		}
		start = i + 1
	}
	return count
}

// populatedLogRange returns the byte range [start, end) of the
// non-0xFF portion of buf (which is exactly the log region). The
// circular buffer is contiguous on disk — the firmware writes head
// forward through it, leaving the tail of unwritten bytes at 0xFF.
// Returns (0, 0) if the buffer is entirely 0xFF.
//
// We trim leading and trailing 0xFF only, since intermediate 0xFF
// bytes can occur naturally (e.g. in binary fields inside a log
// line that hasn't been escaped). This is a simplification that
// works as long as the head cursor hasn't wrapped — once it wraps,
// the populated region is actually the entire buffer and we'd need
// the persisted (head, tail) pair from the cursor sector to know
// where the oldest entry starts.
func populatedLogRange(buf []byte) (int, int) {
	start := 0
	for start < len(buf) && buf[start] == 0xFF {
		start++
	}
	end := len(buf)
	for end > start && buf[end-1] == 0xFF {
		end--
	}
	return start, end
}

// logFFGapThreshold is the minimum run of consecutive 0xFF bytes
// that the log writer leaves between log "blocks" — one block per
// boot session / flush event. Shorter FF runs can occur naturally
// inside a single log line (binary fields), so this lower bound
// keeps a block from splitting on those. Tuned empirically against
// observed dumps; bump if a populated dump shows split sessions.
const logFFGapThreshold = 4

// findLogBlocks returns the [start, end) byte offsets of each
// non-FF "log block" in the 128 KB circular log buffer. Blocks
// are separated by runs of ≥ logFFGapThreshold consecutive 0xFF
// bytes (inter-flush padding); inside a block the firmware writes
// concatenated `text\n` log lines. Blocks that are entirely 0x00
// padding are dropped (pre-initialised flash, not real log data).
//
// The block model coexists with the 128 KB circular-buffer model
// from bleware/src/log_gatt.c: each block is one contiguous write
// burst by the OEM logger between flushes; the buffer itself is
// circular at the BIM level. For host-side inspection of a dump
// the block view matches how a human reads the log ("show me the
// last boot session").
func findLogBlocks(buf []byte) [][2]int {
	var blocks [][2]int
	i, n := 0, len(buf)
	for i < n {
		for i < n && buf[i] == 0xFF {
			i++
		}
		if i >= n {
			break
		}
		start := i
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

// parseLogSelector resolves a comma-separated selector string like
// "0,1,2,last" or "0-4,last" into concrete 0-based indices into a
// slice of `count` log blocks. Supported tokens:
//
//	N        — block index N (negative is rejected)
//	first    — alias for 0
//	last     — count-1
//	N-M      — inclusive range [N, M]; either side may be "first"
//	           or "last"
//
// Duplicates are kept in the order they appear; out-of-range
// indices return an error rather than being silently clamped so
// the user notices a typo.
func parseLogSelector(selector string, count int) ([]int, error) {
	if count <= 0 {
		return nil, fmt.Errorf("no log blocks to select from")
	}
	resolve := func(tok string) (int, error) {
		switch strings.ToLower(strings.TrimSpace(tok)) {
		case "first":
			return 0, nil
		case "last":
			return count - 1, nil
		}
		n, err := strconv.Atoi(strings.TrimSpace(tok))
		if err != nil {
			return 0, fmt.Errorf("invalid log index %q", tok)
		}
		if n < 0 || n >= count {
			return 0, fmt.Errorf("log index %d out of range [0, %d]", n, count-1)
		}
		return n, nil
	}
	var out []int
	for _, part := range strings.Split(selector, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i := strings.Index(part, "-"); i > 0 {
			lo, err := resolve(part[:i])
			if err != nil {
				return nil, err
			}
			hi, err := resolve(part[i+1:])
			if err != nil {
				return nil, err
			}
			if lo > hi {
				lo, hi = hi, lo
			}
			for n := lo; n <= hi; n++ {
				out = append(out, n)
			}
			continue
		}
		n, err := resolve(part)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty selector")
	}
	return out, nil
}

// PrintLogByIndex reads the log region and prints the log blocks
// named by `selector` (e.g. "0", "last", "0,1,2,last", "0-4").
// Always prints the total block count first so the user can see
// what range is valid. Each selected block is printed verbatim,
// preceded by a header showing its absolute flash offset, length,
// and contained line count. Returns an error if the selector is
// malformed or out of range.
func PrintLogByIndex(file *os.File, selector string) error {
	offset := int64(logRegionBase)
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}
	fileSize := stat.Size()
	if offset >= fileSize {
		return fmt.Errorf("log offset 0x%X beyond file size 0x%X", offset, fileSize)
	}
	length := logRegionSize
	if int(fileSize-offset) < length {
		length = int(fileSize - offset)
	}
	buf := readFromFile(file, offset, length)
	if buf == nil {
		return fmt.Errorf("failed to read log region")
	}

	blocks := findLogBlocks(buf)
	if len(blocks) == 0 {
		fmt.Printf("Log region @ 0x%X: no blocks\n", offset)
		return nil
	}
	fmt.Printf("Log region @ 0x%X: %d log blocks (indices 0..%d)\n",
		offset, len(blocks), len(blocks)-1)

	picks, err := parseLogSelector(selector, len(blocks))
	if err != nil {
		return err
	}
	for _, idx := range picks {
		b := blocks[idx]
		blockBytes := buf[b[0]:b[1]]
		lines := countLogLines(blockBytes)
		fmt.Printf("\n--- log [%d] @ 0x%08X (%d bytes, %d entries) ---\n",
			idx, offset+int64(b[0]), b[1]-b[0], lines)
		text := strings.ReplaceAll(string(blockBytes), "\x00", "\n")
		fmt.Println(strings.TrimRight(text, "\n"))
	}
	return nil
}

// countPopulatedBlocks reports how many of the 16-byte entries in
// the log buffer contain at least one non-0xFF byte. Matches what
// `cmd_log_count` (`log_block_count()`) reports over the UART
// monitor.
func countPopulatedBlocks(buf []byte) int {
	count := 0
	for i := 0; i+logEntrySize <= len(buf); i += logEntrySize {
		for _, b := range buf[i : i+logEntrySize] {
			if b != 0xFF {
				count++
				break
			}
		}
	}
	return count
}

// ReadLogsCount summarises the log region. Mirrors what the
// firmware's `cmd_log_count` reports (the "block count" — number of
// populated 16-byte entries) and additionally counts the
// \n-terminated text lines, which is closer to what a human means
// by "how many logs are saved". Layout constants come from
// bleware/src/log_gatt.c (LOG_REGION_BASE / LOG_REGION_SIZE).
func ReadLogsCount(file *os.File) {
	offset := int64(logRegionBase)

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

	length := logRegionSize
	if int(fileSize-offset) < length {
		length = int(fileSize - offset)
	}

	fmt.Printf("Reading %d bytes from log region @ 0x%X (128 KB circular buffer)\n",
		length, offset)

	buf := readFromFile(file, offset, length)
	if buf == nil {
		return
	}

	blocks := findLogBlocks(buf)
	totalLines := 0
	for _, b := range blocks {
		totalLines += countLogLines(buf[b[0]:b[1]])
	}

	fmt.Printf("Log region: %d log blocks containing %d \\n-terminated entries\n",
		len(blocks), totalLines)
	if len(blocks) == 0 {
		return
	}
	for i, b := range blocks {
		lines := countLogLines(buf[b[0]:b[1]])
		fmt.Printf("  log [%d]: 0x%08X - 0x%08X (%d bytes, %d entries)\n",
			i, offset+int64(b[0]), offset+int64(b[1]), b[1]-b[0], lines)
	}
}
