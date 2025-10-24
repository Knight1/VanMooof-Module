package vanmoof

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

var riffMagic = []byte{'R', 'I', 'F', 'F'}

// ExtractWAVFromVMSound extracts WAV data from VM_SOUND by finding RIFF header
func ExtractWAVFromVMSound(vmSoundData []byte) ([]byte, error) {
	// Check for corrupted upload pattern (00 C0 46 C0 repeating)
	if bytes.Contains(vmSoundData, []byte{0x00, 0xC0, 0x46, 0xC0}) {
		return nil, fmt.Errorf("corrupted upload detected - contains failed update data")
	}

	// Find RIFF header in the VM_SOUND data
	riffOffset := bytes.Index(vmSoundData, riffMagic)
	if riffOffset == -1 {
		return nil, fmt.Errorf("no RIFF header found in VM_SOUND data")
	}

	// Return WAV data starting from RIFF header
	return vmSoundData[riffOffset:], nil
}

// ExportVMSoundsAsWAV exports VM_SOUND files as WAV files
func ExportVMSoundsAsWAV(moduleFileName string) error {
	data, err := os.ReadFile(moduleFileName)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	sounds := FindVMSounds(data)
	if len(sounds) == 0 {
		fmt.Println("No VM_SOUND files found")
		return nil
	}

	fmt.Printf("Found %d VM_SOUND files, extracting WAV data...\n", len(sounds))

	baseFileName := filepath.Base(moduleFileName)
	baseFileName = baseFileName[:len(baseFileName)-len(filepath.Ext(baseFileName))]

	for i, sound := range sounds {
		soundData := data[sound.Offset : sound.Offset+sound.Length]

		// Extract WAV data
		wavData, err := ExtractWAVFromVMSound(soundData)
		if err != nil {
			fmt.Printf("Error extracting WAV from sound %d: %v\n", i+1, err)
			continue
		}

		filename := fmt.Sprintf("%s_sound_%02d.wav", baseFileName, i+1)

		err = os.WriteFile(filename, wavData, 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
			continue
		}

		fmt.Printf("  %s (%d bytes)\n", filename, len(wavData))
	}

	return nil
}
