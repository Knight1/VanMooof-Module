package vanmoof

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

var vmSoundMagic = []byte{0x56, 0x4D, 0x5F, 0x53, 0x4F, 0x55, 0x4E, 0x44} // "VM_SOUND"
var vmSoundEnd = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}                     // End marker

func FindVMSounds(data []byte) []VMSound {
	var sounds []VMSound
	offset := 0

	for {
		// Find next VM_SOUND magic
		index := bytes.Index(data[offset:], vmSoundMagic)
		if index == -1 {
			break
		}

		startOffset := offset + index
		// Find end marker after start
		endIndex := bytes.Index(data[startOffset+len(vmSoundMagic):], vmSoundEnd)
		if endIndex == -1 {
			// No end marker found, skip this one
			offset = startOffset + len(vmSoundMagic)
			continue
		}

		endOffset := startOffset + len(vmSoundMagic) + endIndex + len(vmSoundEnd)
		length := endOffset - startOffset

		sounds = append(sounds, VMSound{
			Offset: startOffset,
			Length: length,
		})

		offset = endOffset
	}

	return sounds
}

func ExportVMSounds(moduleFileName string) error {
	file, err := os.Open(moduleFileName)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	data, err := os.ReadFile(moduleFileName)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	sounds := FindVMSounds(data)
	if len(sounds) == 0 {
		fmt.Println("No VM_SOUND files found")
		return nil
	}

	fmt.Printf("Found %d VM_SOUND files, exporting...\n", len(sounds))

	baseFileName := filepath.Base(moduleFileName)
	baseFileName = baseFileName[:len(baseFileName)-len(filepath.Ext(baseFileName))]

	for i, sound := range sounds {
		soundData := data[sound.Offset : sound.Offset+sound.Length]
		filename := fmt.Sprintf("%s_sound_%02d.bin", baseFileName, i+1)

		err := os.WriteFile(filename, soundData, 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
			continue
		}

		fmt.Printf("  %s (%d bytes)\n", filename, sound.Length)
	}

	return nil
}

type VMSound struct {
	Offset int
	Length int
}
