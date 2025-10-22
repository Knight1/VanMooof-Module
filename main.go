package main

import (
	"flag"
)

var (
	moduleFileName  = flag.String("f", "", "Module file name")
	changeUnlockKey = flag.String("u", "", "Change unlock key")
	debugLogging    = flag.Bool("d", false, "Enable debug logging")
	sudo            = flag.Bool("iKnowWhatIAmDoingISwear", false, "Use sudo")
	showBLESecrets  = flag.Bool("show", false, "Show BLE secrets")
	//file            os.File
)

func main() {

	flag.Parse()

	//if *moduleFileName != "" {
	file := loadFile()
	//}

	if *showBLESecrets && *moduleFileName != "" {
		readSecrets(*file)
		readLogs(*file)
	}

	// Extract firmware if module file is provided
	if *moduleFileName != "" {
		checkForFirmware()
	}

	if *changeUnlockKey != "" {
		writeSecrets("unlock", *changeUnlockKey)
	}

	err := file.Close()
	if err != nil {
		return
	}

}
