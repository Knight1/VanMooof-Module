# VanMoof Module Tool Examples

This document contains practical examples of using the VanMoof Module tool for various operations.

## Basic Operations

### Show Dump Content
```console
./cmd -f SPI-Flash.rom -show
Loading File: SPI-Flash.rom
BLE Authentication Key: XXXX
M-ID/M-KEY: XXX
MAC Address: XXXXX
Reading 143360 bytes from offset 0x3fdd000 (file size: 0x4000000)
Found 0 log entries:
Found 23 VM_SOUND files in dump
Found PACK at offset: 0x00080000
PACK Header - Offset: 0x00088A1C, Length: 0x00000140
Extracted PACK to: SPI-Flash.rom.pack (559976 bytes)
Extracting 5 firmware files:
  bleware.bin (181884 bytes)
  mainware.bin (220144 bytes)
  motorware.bin (61720 bytes)
  shifterware.bin (11944 bytes)
  batteryware.bin (83940 bytes)
```

### Export VM_SOUND Files from Dump
```console
./cmd -f SPI-Flash.rom -sounds
Loading File: SPI-Flash.rom
Found 23 VM_SOUND files, exporting...
  SPI-Flash_sound_01.bin (4096 bytes)
  SPI-Flash_sound_02.bin (8192 bytes)
  SPI-Flash_sound_03.bin (2048 bytes)
  ...
```

### Show Logs from Dump
```console
./cmd -f ../rom.rom -logs
Loading File: ../rom.rom
Reading 143360 bytes from offset 0x3fdd000 (file size: 0x4000000)
Found 2 log entries:
Log 1: 
0;39.03;-0.21;0.0;28;32;0.0;4864.3;0;228;1
1723229090 ;21.2;69;27.0;39.04;-0.18;0.0;28;32;0.0;4864.3;0;220;1
1723229119 LiPo state changed to LIPO_DISCHARGING
1723229120 ADC Vbat 22457
1723229120 Set power state to PWR_NORMAL (Current limit: 20.0 A, SOC: -1 %)
1723229120 Set power state to PWR_NORMAL (Current limit: 30.0 A, SOC: 69 %)
1723229120 USER Reset
1723229120 BIKE_RESET
```

## Encryption/Decryption Operations

### Decrypt pack File from API
```console
./cmd -f ../Update1.9.1\ -\ batteryware\ 1.23.1\ x\ mainware\ 1.9.1.pak -decrypt "MANUFACTURING KEY"
Decrypting ../Update1.9.1 - batteryware 1.23.1 x mainware 1.9.1.pak with AES-128 ECB...
File size: 283472 bytes
Decrypted data CRC32: 0x11787DFB
Decrypted PACK saved to: ../Update1.9.1 - batteryware 1.23.1 x mainware 1.9.1_decrypted.pak (283472 bytes)
✓ Decryption successful - valid PACK magic found
PACK Header:
  Magic: PACK
  Directory Offset: 0x000452C8 (283336)
  Directory Length: 0x00000080 (128)
Directory contains 2 entries:
  batteryware.bin (offset: 0x0000000C, length: 87568 bytes)
  mainware.bin (offset: 0x0001561C, length: 195756 bytes)
```

### Encrypt pack File to upload it to the Module via BLE
```console
./cmd -f "Update1.9.1 - batteryware 1.23.1 x mainware 1.9.1_decrypted.pak" -encrypt "MANUFACTURING KEY"
Encrypting Update1.9.1 - batteryware 1.23.1 x mainware 1.9.1_decrypted.pak with AES-128 ECB...
File size: 283472 bytes
Encrypted PACK saved to: Update1.9.1 - batteryware 1.23.1 x mainware 1.9.1_decrypted.pak (283472 bytes)
```

## Security Analysis

### Analyze File Entropy and ECB Patterns (without decryption)
```console
./cmd -f encrypted.pak -entropy
🔍 Entropy Analysis for encrypted.pak:
Shannon entropy: 7.972098 bits/byte (max 8.0)
✅ HIGH ENTROPY: 7.97 bits/byte - likely encrypted data

📊 ECB Block Analysis (16-byte blocks):
Total blocks: 17717
Unique blocks: 15971
Repeated blocks: 2071 (11.69%)
⚠️  HIGH REPETITION WARNING: 11.69% > 10%
   This suggests ECB mode encryption or unencrypted data

🔄 Top repeated blocks:
  1) count=984  hex=2a87ec3f2d7504954583b7b6fc28545f
  2) count= 59  hex=a4bb6c0eb57beaad5255dc8ca7a8d6e1
  ...

📁 File Analysis:
File size: 283472 bytes
File header: "\xb9\xc1kK"
❓ Unknown: Mixed or partially encrypted data
```

### Validate Manufacturing Key Entropy
```console
./cmd -check-key "1234567890ABCDEF1234567890ABCDEF"
🔑 Manufacturing Key Entropy Check:
Key: 1234567890ABCDEF1234567890ABCDEF
Shannon entropy: 3.000000 bits/byte
⚠️  LOW KEY ENTROPY: 3.00 bits/byte - key may be weak
✅ Key validation passed
```

### Weak Key Detection
```console
./cmd -check-key "00000000000000000000000000000000"
🔑 Manufacturing Key Entropy Check:
Key: 00000000000000000000000000000000
Shannon entropy: 0.000000 bits/byte
Key validation failed: WEAK KEY: entropy 0.00 < 3.0 bits/byte - key appears to have patterns
```

### Enhanced Decryption with Entropy Analysis
```console
./cmd -f encrypted.pak -decrypt "4D414E5546414354555249474B455900"
🔍 Entropy Analysis for encrypted.pak:
Shannon entropy: 7.972098 bits/byte (max 8.0)
✅ HIGH ENTROPY: 7.97 bits/byte - likely encrypted data

🔑 Manufacturing Key Entropy Check:
Key: 4D414E5546414354555249474B455900
Shannon entropy: 3.750000 bits/byte
⚠️  LOW KEY ENTROPY: 3.75 bits/byte - key may be weak

Decrypting encrypted.pak with AES-128 ECB...
File size: 283472 bytes
Decrypted data CRC32: 0x04936859

🔍 Entropy Analysis for encrypted.pak (decrypted):
Shannon entropy: 7.971980 bits/byte (max 8.0)
✅ HIGH ENTROPY: 7.97 bits/byte - likely encrypted data

Decrypted PACK saved to: encrypted_decrypted.pak (283472 bytes)
```

## Hardware Operations

### Dump SPI Flash directly from Hardware
```console
sudo ./cmd -dump
🔍 Validating chip compatibility...
✅ Chip validated: MX25L51245G (VanMoof S3 compatible)
🔑 Validating BLE authentication key...
✅ BLE authentication key validated
Dumping 64MB SPI flash to VMES3-1704067200-UNKNOWN_20240101_120000.bin...
Progress: 25.0% (4096/16384 chunks)
Progress: 50.0% (8192/16384 chunks)
Progress: 75.0% (12288/16384 chunks)
Progress: 100.0% (16384/16384 chunks)

Dump completed in 4m32s
File: VMES3-1704067200-UNKNOWN_20240101_120000.bin (64MB)
CRC32: 0x12345678
SHA256: a1b2c3d4e5f6...
✓ Verification passed - file integrity confirmed
✓ Extracted MAC address: F88A5E123456
✓ Renamed file to: VMES3-1704067200-F88A5E123456.bin

🔍 Starting comprehensive dump verification...
📁 Calculating SHA512 from disk file...
📁 Disk SHA512: a1b2c3d4e5f6...
💾 Re-dumping from SPI chip into memory...
🔍 Verification progress: 50.0% (8192/16384 chunks)
🔍 Verification progress: 100.0% (16384/16384 chunks)
💾 Memory SHA512: a1b2c3d4e5f6...
✅ VERIFICATION PASSED: Disk and SPI memory SHA512 match!
```

### Read SPI Flash Chip Information
```console
sudo ./cmd -flash-info
SPI Flash Information:
  Manufacturer: Macronix (0xC2)
  Device: MX25L51245G
  Capacity: 64MB (512Mbit)
  Serial Number: 1234567890ABCDEF
  Unique ID: FEDCBA0987654321
```

### Manual Dump Verification

⚠️  **WARNING**: Keep the module powered down during verification! Powering up the module will write new logs to the flash, causing verification to fail.

```console
sudo ./cmd -f VMES3-1704067200-F88A5E123456.bin -verify
🔍 Starting comprehensive dump verification...
📁 Calculating SHA512 from disk file...
📁 Disk SHA512: a1b2c3d4e5f6789abcdef0123456789abcdef0123456789abcdef0123456789a
💾 Re-dumping from SPI chip into memory...
🔍 Verification progress: 25.0% (4096/16384 chunks)
🔍 Verification progress: 50.0% (8192/16384 chunks)
🔍 Verification progress: 75.0% (12288/16384 chunks)
🔍 Verification progress: 100.0% (16384/16384 chunks)
💾 Memory SHA512: a1b2c3d4e5f6789abcdef0123456789abcdef0123456789abcdef0123456789a
✅ VERIFICATION PASSED: Disk and SPI memory SHA512 match!
```

## Error Handling Examples

### Invalid BLE Key Detection
```console
sudo ./cmd -dump
🔍 Validating chip compatibility...
✅ Chip validated: MX25L51245G (VanMoof S3 compatible)
🔑 Validating BLE authentication key...
⚠️  WARNING: Invalid BLE authentication key detected!
Key: 0000000000000000

❌ INVALID BLE AUTHENTICATION KEY DETECTED

🔧 TROUBLESHOOTING TIPS:
1. Check SPI connections (CLK, MOSI, MISO, CS, GND)
2. Verify 3.3V power supply (not 5V!)
3. Ensure SPI clock speed is not too high (try 1MHz)
4. Check for loose connections or poor contact
5. Module might be in shipping mode - wake it first
6. Try different SPI mode (Mode0/Mode1)

⚠️  Continuing with invalid key may result in corrupted dump!

Do you want to continue anyway? (y/N): N
dump cancelled by user - fix SPI connection and try again
```

### Unsupported Chip Detection
```console
sudo ./cmd -dump
🔍 Validating chip compatibility...
❌ UNSUPPORTED CHIP DETECTED
Manufacturer: 0xEF, Device: 0x4017

⚠️  WARNING: This tool is designed for VanMoof S3 modules only!
Expected: Macronix MX25L51245G (0xC2, 0x201A)
Found: Non-VanMoof Chip (0xEF, 0x4017)

🚫 DUMP ABORTED - Wrong chip type detected
unsupported chip - VanMoof S3 requires MX25L51245G
```

## Hardware Dump with flashrom
```console
# sudo flashrom -p linux_spi:dev=/dev/spidev0.0 -r rom.rom
flashrom v1.2 on Linux 6.1.21+ (armv6l)
flashrom is free software, get the source code at https://flashrom.org

Using clock_gettime for delay loops (clk_id: 1, resolution: 1ns).
Using default 2000kHz clock. Use 'spispeed' parameter to override.
Found Macronix flash chip "MX66L51235F/MX25L51245G" (65536 kB, SPI) on linux_spi.
Reading flash... done.
```