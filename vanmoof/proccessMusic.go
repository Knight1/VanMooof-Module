package vanmoof

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

var vmSoundMagic = []byte{0x56, 0x4D, 0x5F, 0x53, 0x4F, 0x55, 0x4E, 0x44} // "VM_SOUND"
var vmSoundEnd = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}                     // End marker
var vmSoundEndAlt = []byte{0x00, 0x00, 0x00, 0x00, 0x00}                  // Alternative end marker

// VM_SOUND audio storage layout per bleware/src/monitor/cmd_audio.c:
//
//   base   = 0x00200000  (slot 0)
//   stride = 0x00080000  (512 KiB per slot)
//   count  = 0x7B        (123 slots, indices 0..0x7A inclusive)
//
// The 123 slots span 0x200000..0x3F80000 (63 MiB). Earlier matches
// of the "VM_SOUND" byte sequence are firmware-code references
// (mainware/bleware string pool), not audio files, and must be
// ignored — otherwise -verify reports phantom VM_SOUND hits inside
// the PACK firmware region.
const (
	vmSoundFirstOffset = 0x00200000
	vmSoundAlignment   = 0x00080000
	vmSoundSlotCount   = 0x7B
	vmSoundRegionEnd   = vmSoundFirstOffset + vmSoundSlotCount*vmSoundAlignment
)

func FindVMSounds(data []byte) []VMSound {
	var sounds []VMSound
	offset := vmSoundFirstOffset
	if offset > len(data) {
		return sounds
	}

	// Hard upper bound: the audio region ends at slot 123. Anything
	// past that belongs to the logs sector or trailing flash.
	scanEnd := vmSoundRegionEnd
	if scanEnd > len(data) {
		scanEnd = len(data)
	}

	for offset < scanEnd {
		// Find next VM_SOUND magic
		index := bytes.Index(data[offset:scanEnd], vmSoundMagic)
		if index == -1 {
			break
		}

		startOffset := offset + index

		// Real audio files sit at sector-aligned offsets. Anything
		// else is a firmware-string false positive — skip past it.
		if startOffset%vmSoundAlignment != 0 {
			offset = startOffset + len(vmSoundMagic)
			continue
		}

		// Find end marker after start (try both patterns)
		endIndex := bytes.Index(data[startOffset+len(vmSoundMagic):], vmSoundEnd)
		endIndexAlt := bytes.Index(data[startOffset+len(vmSoundMagic):], vmSoundEndAlt)

		// Use whichever end marker is found first (or closest)
		if endIndex == -1 && endIndexAlt == -1 {
			// No end marker found, skip this one
			offset = startOffset + len(vmSoundMagic)
			continue
		}

		if endIndex == -1 || (endIndexAlt != -1 && endIndexAlt < endIndex) {
			endIndex = endIndexAlt
		}

		endOffset := startOffset + len(vmSoundMagic) + endIndex + 5
		length := endOffset - startOffset

		sounds = append(sounds, VMSound{
			Slot:   (startOffset - vmSoundFirstOffset) / vmSoundAlignment,
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
	defer func() { _ = file.Close() }()

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

	for _, sound := range sounds {
		soundData := data[sound.Offset : sound.Offset+sound.Length]
		filename := fmt.Sprintf("%s_sound_slot%02d.bin", baseFileName, sound.Slot)

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
	// Slot is the firmware-level index passed to audio_play / used
	// by audio_upload <index> — derived from (Offset - 0x200000) /
	// 0x80000. Range [0, 0x7A] inclusive.
	Slot   int
	Offset int
	Length int
}
