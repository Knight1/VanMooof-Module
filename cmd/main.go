package main

import (
	"VanMooof-Module/vanmoof"
	"flag"
)

var (
	ModuleFileName  = flag.String("f", "", "Module file name")
	changeUnlockKey = flag.String("u", "", "Change unlock key")
	showBLESecrets  = flag.Bool("show", false, "Show BLE secrets")
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

	// Extract firmware if module file is provided
	if *ModuleFileName != "" {
		vanmoof.CheckForFirmware(ModuleFileName)
	}

	if *changeUnlockKey != "" {
		vanmoof.WriteSecrets("unlock", *changeUnlockKey)
	}

	err := file.Close()
	if err != nil {
		return
	}

}
