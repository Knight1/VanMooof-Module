package main

import (
	"VanMooof-Module/vanmoof"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	ModuleFileName  = flag.String("f", "", "Module file name")
	changeUnlockKey = flag.String("u", "", "Change unlock key")
	showBLESecrets  = flag.Bool("show", false, "Show BLE secrets")
	showLogs        = flag.Bool("logs", false, "Show logs only")
	extractPack     = flag.Bool("pack", false, "Extract PACK file only (without extracting individual firmware files)")
	exportSounds    = flag.Bool("sounds", false, "Export VM_SOUND files from SPI dump")
	uploadPack      = flag.String("upload", "", "Upload PACK file via Y-Modem (specify PACK file path)")
	serialPort      = flag.String("port", "", "Serial port for Y-Modem upload (auto-detect if empty)")
	listPorts       = flag.Bool("list-ports", false, "List available serial ports")
	decryptPack     = flag.String("decrypt", "", "Decrypt PACK file with AES ECB (specify key in hex)")
	encryptPack     = flag.String("encrypt", "", "Encrypt PACK file with AES ECB (specify key in hex)")
	checkEntropy    = flag.Bool("entropy", false, "Analyze file entropy and ECB patterns without decryption")
	checkKey        = flag.String("check-key", "", "Validate manufacturing key entropy (specify key in hex)")
	verifyDump      = flag.Bool("verify", false, "Verify SPI dump data coverage")
	showExtra       = flag.Bool("extra", false, "Show unaccounted data regions (use with -verify)")
	dumpFlash       = flag.String("dump", "", "Dump SPI flash to file (optional format: MAC,FRAME or MAC or empty for auto-detect)")
	flashInfo       = flag.Bool("flash-info", false, "Read SPI flash chip information and serial number")
	sudoFlag        = flag.Bool("sudo", false, "Enable SPI hardware access (required for dump and flash-info)")
	//file            os.File
)

func main() {
	flag.Usage = func() {
		log.Printf("VanMoof Module Tool - PACK Upload and SPI Flash Analysis\n\n")
		log.Printf("Usage: %s [options]\n\n", os.Args[0])
		log.Printf("Options:\n")
		flag.PrintDefaults()
		log.Printf("\nExamples:\n")
		log.Printf("  %s -list-ports                    # List available serial ports\n", os.Args[0])
		log.Printf("  %s -upload pack.bin               # Upload PACK file (115200 baud)\n", os.Args[0])
		log.Printf("  %s -f dump.rom -show              # Analyze SPI flash dump\n", os.Args[0])
		log.Printf("  %s -f dump.rom -logs              # Show logs only\n", os.Args[0])
		log.Printf("  %s -f dump.rom -pack              # Extract PACK from dump\n", os.Args[0])
		log.Printf("  %s -f dump.rom -sounds            # Export VM_SOUND files\n", os.Args[0])
		log.Printf("  %s -f dump.rom -verify            # Verify data coverage\n", os.Args[0])
		log.Printf("  %s -f dump.rom -verify -extra     # Show unaccounted regions\n", os.Args[0])
		log.Printf("  %s -f pack.bin -decrypt KEY       # Decrypt PACK file with AES ECB\n", os.Args[0])
		log.Printf("  %s -f pack.bin -encrypt KEY       # Encrypt PACK file with AES ECB\n", os.Args[0])
		log.Printf("  %s -f pack.bin -entropy           # Analyze file entropy without decryption\n", os.Args[0])
		log.Printf("  %s -check-key KEY                 # Validate manufacturing key entropy\n", os.Args[0])
		log.Printf("  %s -dump MAC,FRAME                # Dump SPI flash (e.g. F88A5E123456,2043531337)\n", os.Args[0])
		log.Printf("  %s -dump -sudo                    # Dump SPI flash (auto-detect MAC from dump)\n", os.Args[0])
		log.Printf("  %s -flash-info -sudo              # Read SPI flash chip info and serial number\n", os.Args[0])
	}

	flag.Parse()

	// List available serial ports (no file required)
	if *listPorts {
		ports, err := vanmoof.ListSerialPorts()
		if err != nil {
			fmt.Printf("Error listing ports: %v\n", err)
		} else {
			fmt.Println("Available serial ports:")
			for _, port := range ports {
				fmt.Printf("  %s\n", port)
			}
		}
		return
	}

	// Dump SPI flash (no file required) - check if dump flag was used
	dumpFlagUsed := false
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-dump") {
			dumpFlagUsed = true
			break
		}
	}
	if dumpFlagUsed {
		var macAddress, frameNumber string

		if *dumpFlash == "" {
			// Use current date/time as fallback
			now := time.Now()
			macAddress = fmt.Sprintf("UNKNOWN_%s", now.Format("20060102_150405"))
			frameNumber = fmt.Sprintf("%d", now.Unix())
		} else {
			parts := strings.Split(*dumpFlash, ",")
			if len(parts) == 1 {
				// Only MAC provided, use timestamp for frame
				macAddress = strings.TrimSpace(parts[0])
				frameNumber = fmt.Sprintf("%d", time.Now().Unix())
			} else if len(parts) == 2 {
				macAddress = strings.TrimSpace(parts[0])
				frameNumber = strings.TrimSpace(parts[1])
			} else {
				fmt.Println("Invalid dump format. Use: MAC,FRAME or MAC or leave empty")
				os.Exit(1)
			}
		}

		err := vanmoof.DumpFlash(macAddress, frameNumber, *sudoFlag)
		if err != nil {
			fmt.Printf("Flash dump failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Read SPI flash information (no file required)
	if *flashInfo {
		info, err := vanmoof.ReadFlashInfo(*sudoFlag)
		if err != nil {
			fmt.Printf("Failed to read flash info: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(info.String())
		return
	}

	// Upload PACK file via Y-Modem
	if *uploadPack != "" {
		if _, err := os.Stat(*uploadPack); os.IsNotExist(err) {
			fmt.Printf("Pack file not provided or does not exist: %s\n", *uploadPack)
			os.Exit(1)
		}

		port := *serialPort
		if port == "" {
			port = vanmoof.GetDefaultSerialPort()
			fmt.Printf("Using default serial port: %s\n", port)
		}

		if err := vanmoof.ValidateSerialPort(port); err != nil {
			fmt.Printf("Invalid serial port: %v\n", err)
			os.Exit(1)
		}

		err := vanmoof.UploadPACK(*uploadPack, port, 115200)
		if err != nil {
			fmt.Printf("Upload failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Decrypt PACK file
	if *decryptPack != "" {
		if *ModuleFileName == "" {
			fmt.Println("Pack file path required. Use -f PACKFILE")
			os.Exit(1)
		}
		if _, err := os.Stat(*ModuleFileName); os.IsNotExist(err) {
			fmt.Printf("Pack file does not exist: %s\n", *ModuleFileName)
			os.Exit(1)
		}
		err := vanmoof.DecryptPack(*ModuleFileName, *decryptPack)
		if err != nil {
			fmt.Printf("Decryption failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Encrypt PACK file
	if *encryptPack != "" {
		if *ModuleFileName == "" {
			fmt.Println("Pack file path required. Use -f PACKFILE")
			os.Exit(1)
		}
		if _, err := os.Stat(*ModuleFileName); os.IsNotExist(err) {
			fmt.Printf("Pack file does not exist: %s\n", *ModuleFileName)
			os.Exit(1)
		}
		err := vanmoof.EncryptPack(*ModuleFileName, *encryptPack)
		if err != nil {
			fmt.Printf("Encryption failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check manufacturing key entropy (no file required)
	if *checkKey != "" {
		err := vanmoof.CheckKeyEntropy(*checkKey)
		if err != nil {
			fmt.Printf("Key validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Key validation passed")
		return
	}

	// Analyze file entropy (requires file)
	if *checkEntropy {
		if *ModuleFileName == "" {
			fmt.Println("File path required. Use -f FILE")
			os.Exit(1)
		}
		if _, err := os.Stat(*ModuleFileName); os.IsNotExist(err) {
			fmt.Printf("File does not exist: %s\n", *ModuleFileName)
			os.Exit(1)
		}
		data, err := os.ReadFile(*ModuleFileName)
		if err != nil {
			fmt.Printf("Failed to read file: %v\n", err)
			os.Exit(1)
		}
		vanmoof.AnalyzeFileEntropy(*ModuleFileName, data)
		return
	}

	// File operations require a module file
	var file *os.File
	if *ModuleFileName != "" {
		file = vanmoof.LoadFile(ModuleFileName)
	} else {
		// Check if any file-dependent operations are requested
		if *showBLESecrets || *showLogs || *changeUnlockKey != "" || *exportSounds || *verifyDump || *checkEntropy {
			fmt.Println("File path required. Use -f FILE")
			os.Exit(1)
		}
		// If no operations specified, show usage
		if flag.NFlag() == 0 {
			flag.Usage()
			os.Exit(1)
		}
		return
	}

	if *showLogs {
		vanmoof.ReadLogs(file)
		return
	}

	if *exportSounds {
		err := vanmoof.ExportVMSounds(*ModuleFileName)
		if err != nil {
			fmt.Printf("Error exporting sounds: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *verifyDump {
		err := vanmoof.VerifyDump(*ModuleFileName, *showExtra)
		if err != nil {
			fmt.Printf("Error verifying dump: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *showBLESecrets {
		vanmoof.ReadSecrets(file)
		vanmoof.ReadLogsCount(file)
		// Always check for firmware to show PACK contents when using -show
		vanmoof.CheckForFirmware(ModuleFileName, *extractPack)
	} else {
		// Only check for firmware if not showing secrets (to avoid duplicate output)
		vanmoof.CheckForFirmware(ModuleFileName, *extractPack)
	}

	if *changeUnlockKey != "" {
		vanmoof.WriteSecrets("unlock", *changeUnlockKey)
	}

	if file != nil {
		_ = file.Close()
	}

}
