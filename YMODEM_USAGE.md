# Y-Modem PACK Upload Usage

## Overview
This tool now supports uploading PACK files to VanMoof modules via Y-Modem protocol over USB UART adapters.

## Features
- **PACK File Validation**: Ensures the file is a valid PACK file and not an SPI dump
- **Header Validation**: Verifies PACK magic bytes, offset, and length
- **Size Check**: Rejects files larger than 2MB
- **Directory Validation**: Checks PACK directory structure and entries
- **Y-Modem Upload**: Transfers validated PACK files via serial connection at 115200 baud
- **Cross-Platform**: Works on Windows, macOS, and Linux

## Usage Examples

### List Available Serial Ports
```bash
./cmd -list-ports
```

### Upload PACK File
```bash
# Upload with auto-detected default port (always 115200 baud)
./cmd -upload "pack.bin"

# Windows - Upload with custom port
./cmd.exe -upload "pack.bin" -port COM5

# macOS - Upload with USB serial adapter
./cmd -upload "pack.bin" -port /dev/tty.usbserial-0001

# Linux - Upload with USB serial adapter
./cmd -upload "pack.bin" -port /dev/ttyUSB0

# Upload PACK file
./cmd -upload "packFile - mainware 1.09.03 - batteryware 1.17 - bleware 2.4.01 - motorware S.0.00.22 - shifterware 0.237.bin"
```

## Hardware Setup
1. Connect USB UART adapter to debug port
2. Use JTAG pinout: Black=GND, Green=TX, Orange=RX, Yellow=NC

## Platform-Specific Serial Ports
- **Windows**: COM1, COM2, COM3, etc.
- **macOS**: /dev/tty.usbserial-*, /dev/tty.usbmodem-*, /dev/tty.SLAB_USBtoUART*
- **Linux**: /dev/ttyUSB*, /dev/ttyACM*, /dev/ttyS*

## Default Serial Ports
- **Windows**: COM3
- **macOS**: /dev/tty.usbserial-0001
- **Linux**: /dev/ttyUSB0

## PACK File Requirements
- Must start with "PACK" magic bytes
- File size must be less than 2MB
- Must contain valid directory structure
- Directory entries must have valid filenames

## Error Handling
The tool will reject:
- Files larger than 2MB (exceeds Chip limits)
- Files without PACK magic header
- Files with invalid directory structure
- Files with corrupted headers

## BLE Shell Upload Process
1. Enter BLE debug shell: `bledebug` in main console
2. Delete existing PACK: `pack-delete`
3. Start upload: `pack-upload`
4. Start tool to Transfer Pack
5. Process PACK: `pack-process`
6. System will reboot with new firmware
