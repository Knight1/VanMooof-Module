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
	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Skip RIFF header (12 bytes)
	file.Seek(12, 0)

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
			binary.Read(file, binary.LittleEndian, &audioFormat)
			binary.Read(file, binary.LittleEndian, &info.Channels)
			binary.Read(file, binary.LittleEndian, &info.SampleRate)
			file.Seek(6, 1) // Skip ByteRate and BlockAlign
			binary.Read(file, binary.LittleEndian, &info.BitsPerSample)
			break
		} else {
			file.Seek(int64(chunkSize), 1)
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
	data, err := os.ReadFile(moduleFileName)
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
			fmt.Printf("Sound %02d: Error - %s\n", i+1, err.Error())
			continue
		}

		// Write temp file for analysis
		tempDir := os.TempDir()
		tempFile := filepath.Join(tempDir, fmt.Sprintf("temp_%02d.wav", i+1))
		os.WriteFile(tempFile, wavData, 0644)

		info, err := AnalyzeWAV(tempFile)
		os.Remove(tempFile)

		if err != nil {
			fmt.Printf("Sound %02d: Analysis error - %v\n", i+1, err)
			continue
		}

		fmt.Printf("Sound %02d: %d Hz, %d-bit, %d ch, %.2fs, %d bytes\n",
			i+1, info.SampleRate, info.BitsPerSample, info.Channels, info.Duration, len(wavData))
	}

	return nil
}
