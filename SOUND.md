# VanMoof VM_SOUND Analysis

## Overview

VanMoof bikes store audio files in a custom `VM_SOUND` format within the SPI flash memory. These files contain the various sounds used by the bike including bells, alerts, and system notifications.

## VM_SOUND Format Structure

```
VM_SOUND Header: 56 4D 5F 53 4F 55 4E 44 (8 bytes - "VM_SOUND")
Header Data:     FF FF FF FF 01 58 58 58 58 00 58 58 58 58 58 58 6C 54 00 00 (20 bytes)
WAV Data:        52 49 46 46... (RIFF header + audio data)
End Marker:      FF FF FF FF FF (5 bytes) OR 00 00 00 00 00 (5 bytes)
```

### Header Breakdown

```
56 4D 5F 53 4F 55 4E 44  - "VM_SOUND" magic (8 bytes)
FF FF FF FF               - Unknown/padding (4 bytes)
01                        - Version/type flag (1 byte)
58 58 58 58               - Unknown data (4 bytes)
00                        - Separator/null (1 byte)
58 58 58 58 58 58         - Unknown data (6 bytes)
6C 54 00 00               - WAV data size (4 bytes, little endian)
52 49 46 46               - RIFF header start (WAV data begins)
```

### Size Field Calculation

The 4-byte size field at offset 24-27 indicates WAV data length:
- `6C 54 00 00` = 0x0000546C = 21,612 bytes
- `0A 1F 04 00` = 0x00041F0A = 270,090 bytes

To calculate: WAV_size_decimal → hex → reverse_bytes (little endian)

The VM_SOUND format is a simple header prepended to standard WAV files, not a proprietary wrapper.

## Audio Properties

Most VanMoof sound files use:
- **Sample Rate**: 44.1 kHz (CD quality)
- **Bit Depth**: 16-bit PCM
- **Channels**: Mono (1 channel)
- **Format**: Standard WAV/RIFF format embedded within VM_SOUND wrapper

## Tool Usage

### Export VM_SOUND Files (Binary)
```bash
./cmd -f dump.rom -sounds
```
Exports raw VM_SOUND files with headers intact.

### Extract WAV Files
```bash
./cmd -f dump.rom -wav
```
Extracts playable WAV files by finding RIFF headers within VM_SOUND data.

### Analyze Audio Properties
```bash
./cmd -f dump.rom -analyze-wav
```
Shows detailed audio information:
- Sample rate, bit depth, channels
- Duration and file size
- Detects corrupted files

## Common Issues

### Corrupted Upload Data
Failed firmware updates can leave corrupted VM_SOUND entries containing debug text instead of audio:

**Pattern**: `00 C0 46 C0` repeating throughout the data
**Content**: ASCII text like "source/tasks/audiotask.c" and error messages
**Detection**: Tool automatically identifies and warns about these corrupted entries

### Multiple End Markers
Different firmware versions use different end markers:
- Older versions: `FF FF FF FF FF`
- Newer versions: `00 00 00 00 00`

The tool automatically detects both patterns.

## Memory Layout

VM_SOUND files are typically located at these SPI flash offsets:
```
0x0280000 - VM_SOUND #1
0x0300000 - VM_SOUND #2
0x0380000 - VM_SOUND #3
...
(Additional sounds at various offsets)
```

## Sound Types

Based on analysis, VanMoof bikes typically contain:
1. **Bell sounds** (various tones and melodies)
2. **System alerts** (lock/unlock confirmations)
3. **Error notifications** (warning beeps)
4. **Status sounds** (power on/off, charging)

Duration ranges from ~1.3 seconds to ~20 seconds depending on the sound type.

## Technical Notes

- VM_SOUND is a simple header format prepended to standard WAV files
- The actual audio data is standard RIFF/WAV format
- Some entries may be placeholders or corrupted from failed updates
- Audio quality is consistent across all valid sound files (44.1kHz/16-bit/mono)