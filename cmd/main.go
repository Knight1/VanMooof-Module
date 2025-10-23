package main

import (
	"VanMooof-Module/vanmoof"
	"flag"
)

var (
	ModuleFileName  = flag.String("f", "", "Module file name")
	changeUnlockKey = flag.String("u", "", "Change unlock key")
	showBLESecrets  = flag.Bool("show", false, "Show BLE secrets")
	extractPack     = flag.Bool("pack", false, "Extract PACK file only (without extracting individual firmware files)")
	//file            os.File
)

func main() {

	flag.Parse()

	//if *ModuleFileName != "" {
	file := vanmoof.LoadFile(ModuleFileName)
	//}

	if *showBLESecrets && *ModuleFileName != "" {
		vanmoof.ReadSecrets(file)
		vanmoof.ReadLogs(file)
	}

	// Always check for firmware to show PACK contents
	if *ModuleFileName != "" {
		vanmoof.CheckForFirmware(ModuleFileName, *extractPack)
	}

	if *changeUnlockKey != "" {
		vanmoof.WriteSecrets("unlock", *changeUnlockKey)
	}

	err := file.Close()
	if err != nil {
		return
	}

}
