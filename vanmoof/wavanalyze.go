package vanmoof

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

type WAVInfo struct {
	SampleRate    uint32
	Channels      uint16
	BitsPerSample uint16
	Duration      float64
	FileSize      int64
}

func AnalyzeWAV(filename string) (*WAVInfo, error) {
	cleanPath := filepath.Clean(filename)
	if !filepath.IsAbs(cleanPath) {
		return nil, fmt.Errorf("path must be absolute")
	}
	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Skip RIFF header (12 bytes)
	if _, err := file.Seek(12, 0); err != nil {
		return nil, err
	}

	var info WAVInfo
	info.FileSize = stat.Size()

	// Read fmt chunk
	for {
		var chunkID [4]byte
		var chunkSize uint32

		if err := binary.Read(file, binary.LittleEndian, &chunkID); err != nil {
			break
		}
		if err := binary.Read(file, binary.LittleEndian, &chunkSize); err != nil {
			break
		}

		if string(chunkID[:]) == "fmt " {
			var audioFormat uint16
			if err := binary.Read(file, binary.LittleEndian, &audioFormat); err != nil {
				return nil, err
			}
			if err := binary.Read(file, binary.LittleEndian, &info.Channels); err != nil {
				return nil, err
			}
			if err := binary.Read(file, binary.LittleEndian, &info.SampleRate); err != nil {
				return nil, err
			}
			if _, err := file.Seek(6, 1); err != nil {
				return nil, err
			}
			if err := binary.Read(file, binary.LittleEndian, &info.BitsPerSample); err != nil {
				return nil, err
			}
			break
		} else {
			if _, err := file.Seek(int64(chunkSize), 1); err != nil {
				return nil, err
			}
		}
	}

	// Calculate duration
	bytesPerSample := int64(info.Channels) * int64(info.BitsPerSample) / 8
	if bytesPerSample > 0 && info.SampleRate > 0 {
		dataSize := info.FileSize - 44 // Approximate data size
		info.Duration = float64(dataSize) / float64(bytesPerSample) / float64(info.SampleRate)
	}

	return &info, nil
}

func AnalyzeVMSoundWAVs(moduleFileName string) error {
	cleanPath := filepath.Clean(moduleFileName)
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("path must be absolute")
	}
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	sounds := FindVMSounds(data)
	if len(sounds) == 0 {
		fmt.Println("No VM_SOUND files found")
		return nil
	}

	fmt.Printf("Analyzing %d VM_SOUND files...\n\n", len(sounds))

	baseFileName := filepath.Base(moduleFileName)
	baseFileName = baseFileName[:len(baseFileName)-len(filepath.Ext(baseFileName))]

	for i, sound := range sounds {
		soundData := data[sound.Offset : sound.Offset+sound.Length]

		wavData, err := ExtractWAVFromVMSound(soundData)
		if err != nil {
			fmt.Printf("Sound %02d: Error - failed to extract WAV\n", i+1)
			continue
		}

		// Write temp file for analysis
		tempDir := os.TempDir()
		tempFile := filepath.Join(tempDir, fmt.Sprintf("temp_%02d.wav", i+1))
		if err := os.WriteFile(tempFile, wavData, 0644); err != nil {
			fmt.Printf("Sound %02d: Failed to write temp file\n", i+1)
			continue
		}

		info, err := AnalyzeWAV(tempFile)
		if removeErr := os.Remove(tempFile); removeErr != nil {
			fmt.Printf("Warning: Failed to remove temp file\n")
		}

		if err != nil {
			fmt.Printf("Sound %02d: Analysis error\n", i+1)
			continue
		}

		fmt.Printf("Sound %02d: %d Hz, %d-bit, %d ch, %.2fs, %d bytes\n",
			i+1, info.SampleRate, info.BitsPerSample, info.Channels, info.Duration, len(wavData))
	}

	return nil
}
