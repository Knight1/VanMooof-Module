package main

import (
	"VanMooof-Module/vanmoof"
	"flag"
	"fmt"
	"os"
)

var (
	ModuleFileName  = flag.String("f", "", "Module file name")
	changeUnlockKey = flag.String("u", "", "Change unlock key")
	showBLESecrets  = flag.Bool("show", false, "Show BLE secrets")
	extractPack     = flag.Bool("pack", false, "Extract PACK file only (without extracting individual firmware files)")
	uploadPack      = flag.String("upload", "", "Upload PACK file via Y-Modem (specify PACK file path)")
	serialPort      = flag.String("port", "", "Serial port for Y-Modem upload (auto-detect if empty)")
	listPorts       = flag.Bool("list-ports", false, "List available serial ports")
	//file            os.File
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "VanMoof Module Tool - PACK Upload and SPI Flash Analysis\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -list-ports                    # List available serial ports\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -upload pack.bin               # Upload PACK file (115200 baud)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -f dump.rom -show              # Analyze SPI flash dump\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -f dump.rom -pack              # Extract PACK from dump\n", os.Args[0])
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

	// Upload PACK file via Y-Modem (no file required)
	if *uploadPack != "" {
		port := *serialPort
		if port == "" {
			port = vanmoof.GetDefaultSerialPort()
			fmt.Printf("Using default serial port: %s\n", port)
		}

		// Validate serial port format
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

	// File operations require a module file
	var file *os.File
	if *ModuleFileName != "" {
		file = vanmoof.LoadFile(ModuleFileName)
	} else {
		// Check if any file-dependent operations are requested
		if *showBLESecrets || *changeUnlockKey != "" {
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

	if *showBLESecrets {
		vanmoof.ReadSecrets(file)
		vanmoof.ReadLogs(file)
	}

	// Always check for firmware to show PACK contents
	vanmoof.CheckForFirmware(ModuleFileName, *extractPack)

	if *changeUnlockKey != "" {
		vanmoof.WriteSecrets("unlock", *changeUnlockKey)
	}

	if file != nil {
		err := file.Close()
		if err != nil {
			return
		}
	}

}
