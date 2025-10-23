package vanmoof

import (
	"encoding/hex"
	"fmt"
	"math"
	"sort"
)

// EntropyResult contains entropy analysis results
type EntropyResult struct {
	ShannonEntropy    float64
	TotalBlocks       int
	UniqueBlocks      int
	RepeatedBlocks    int
	RepetitionPercent float64
	TopRepeatedBlocks []BlockCount
}

// BlockCount represents a block and its occurrence count
type BlockCount struct {
	Block []byte
	Count int
}

// CalculateEntropy computes Shannon entropy and ECB repetition analysis
func CalculateEntropy(data []byte) EntropyResult {
	if len(data) == 0 {
		return EntropyResult{}
	}

	// Calculate Shannon entropy
	byteCounts := make(map[byte]int)
	for _, b := range data {
		byteCounts[b]++
	}

	entropy := shannonEntropy(byteCounts, len(data))

	// AES block analysis (16 bytes)
	const blockSize = 16
	blockCounts := make(map[string]int)

	for i := 0; i+blockSize <= len(data); i += blockSize {
		key := string(data[i : i+blockSize])
		blockCounts[key]++
	}

	totalBlocks := 0
	for _, c := range blockCounts {
		totalBlocks += c
	}

	uniqueBlocks := len(blockCounts)
	repeatedBlocks := 0

	for _, c := range blockCounts {
		if c > 1 {
			repeatedBlocks += c
		}
	}

	repetitionPercent := 0.0
	if totalBlocks > 0 {
		repetitionPercent = 100.0 * float64(repeatedBlocks) / float64(totalBlocks)
	}

	// Get top repeated blocks
	var topBlocks []BlockCount
	for k, v := range blockCounts {
		if v > 1 {
			topBlocks = append(topBlocks, BlockCount{
				Block: []byte(k),
				Count: v,
			})
		}
	}

	sort.Slice(topBlocks, func(i, j int) bool {
		return topBlocks[i].Count > topBlocks[j].Count
	})

	// Limit to top 10
	if len(topBlocks) > 10 {
		topBlocks = topBlocks[:10]
	}

	return EntropyResult{
		ShannonEntropy:    entropy,
		TotalBlocks:       totalBlocks,
		UniqueBlocks:      uniqueBlocks,
		RepeatedBlocks:    repeatedBlocks,
		RepetitionPercent: repetitionPercent,
		TopRepeatedBlocks: topBlocks,
	}
}

// shannonEntropy calculates Shannon entropy in bits per byte
func shannonEntropy(counts map[byte]int, total int) float64 {
	h := 0.0
	for _, c := range counts {
		if c == 0 {
			continue
		}
		p := float64(c) / float64(total)
		h += -p * math.Log2(p)
	}
	return h
}

// PrintEntropyAnalysis prints detailed entropy analysis
func PrintEntropyAnalysis(result EntropyResult, filename string) {
	fmt.Printf("\n🔍 Entropy Analysis for %s:\n", filename)
	fmt.Printf("Shannon entropy: %.6f bits/byte (max 8.0)\n", result.ShannonEntropy)

	if result.ShannonEntropy < 7.0 {
		fmt.Printf("⚠️  LOW ENTROPY WARNING: %.2f < 7.0 bits/byte\n", result.ShannonEntropy)
		fmt.Printf("   This suggests the data may be unencrypted or poorly encrypted\n")
	} else if result.ShannonEntropy >= 7.8 {
		fmt.Printf("✅ HIGH ENTROPY: %.2f bits/byte - likely encrypted data\n", result.ShannonEntropy)
	} else {
		fmt.Printf("⚠️  MEDIUM ENTROPY: %.2f bits/byte - encryption quality uncertain\n", result.ShannonEntropy)
	}

	fmt.Printf("\n📊 ECB Block Analysis (16-byte blocks):\n")
	fmt.Printf("Total blocks: %d\n", result.TotalBlocks)
	fmt.Printf("Unique blocks: %d\n", result.UniqueBlocks)
	fmt.Printf("Repeated blocks: %d (%.2f%%)\n", result.RepeatedBlocks, result.RepetitionPercent)

	if result.RepetitionPercent > 10.0 {
		fmt.Printf("⚠️  HIGH REPETITION WARNING: %.2f%% > 10%%\n", result.RepetitionPercent)
		fmt.Printf("   This suggests ECB mode encryption or unencrypted data\n")
	} else if result.RepetitionPercent > 5.0 {
		fmt.Printf("⚠️  MEDIUM REPETITION: %.2f%% - possible ECB patterns\n", result.RepetitionPercent)
	} else {
		fmt.Printf("✅ LOW REPETITION: %.2f%% - good encryption randomness\n", result.RepetitionPercent)
	}

	if len(result.TopRepeatedBlocks) > 0 {
		fmt.Printf("\n🔄 Top repeated blocks:\n")
		for i, block := range result.TopRepeatedBlocks {
			fmt.Printf(" %2d) count=%3d  hex=%s\n", i+1, block.Count, hex.EncodeToString(block.Block))
		}
	}
}

// CheckKeyEntropy validates manufacturing key entropy
func CheckKeyEntropy(keyHex string) error {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("invalid hex key: %v", err)
	}

	if len(key) != 16 {
		return fmt.Errorf("invalid key length: %d bytes (expected 16)", len(key))
	}

	result := CalculateEntropy(key)

	fmt.Printf("\n🔑 Manufacturing Key Entropy Check:\n")
	fmt.Printf("Key: %s\n", keyHex)
	fmt.Printf("Shannon entropy: %.6f bits/byte\n", result.ShannonEntropy)

	// Check for weak keys
	if result.ShannonEntropy < 3.0 {
		return fmt.Errorf("WEAK KEY: entropy %.2f < 3.0 bits/byte - key appears to have patterns", result.ShannonEntropy)
	} else if result.ShannonEntropy < 4.0 {
		fmt.Printf("⚠️  LOW KEY ENTROPY: %.2f bits/byte - key may be weak\n", result.ShannonEntropy)
	} else {
		fmt.Printf("✅ GOOD KEY ENTROPY: %.2f bits/byte\n", result.ShannonEntropy)
	}

	// Check for repeated bytes
	byteCounts := make(map[byte]int)
	for _, b := range key {
		byteCounts[b]++
	}

	maxCount := 0
	var repeatedByte byte
	for b, count := range byteCounts {
		if count > maxCount {
			maxCount = count
			repeatedByte = b
		}
	}

	if maxCount > 8 {
		return fmt.Errorf("WEAK KEY: byte 0x%02X appears %d times (>50%% of key)", repeatedByte, maxCount)
	} else if maxCount > 4 {
		fmt.Printf("⚠️  KEY WARNING: byte 0x%02X appears %d times\n", repeatedByte, maxCount)
	}

	return nil
}

// AnalyzeFileEntropy analyzes a file without decrypting it
func AnalyzeFileEntropy(filename string, data []byte) {
	result := CalculateEntropy(data)
	PrintEntropyAnalysis(result, filename)

	// Additional file-specific analysis
	fmt.Printf("\n📁 File Analysis:\n")
	fmt.Printf("File size: %d bytes\n", len(data))

	if len(data) >= 4 {
		magic := string(data[0:4])
		fmt.Printf("File header: %q\n", magic)

		if magic == "PACK" {
			fmt.Printf("✅ Detected: Unencrypted PACK file\n")
		} else if result.ShannonEntropy >= 7.8 && result.RepetitionPercent < 5.0 {
			fmt.Printf("🔒 Likely: Encrypted file (high entropy, low repetition)\n")
		} else if result.ShannonEntropy < 6.0 {
			fmt.Printf("📄 Likely: Unencrypted or compressed data\n")
		} else {
			fmt.Printf("❓ Unknown: Mixed or partially encrypted data\n")
		}
	}
}
