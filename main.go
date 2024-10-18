package main

import (
	"flag"
)

var (
	moduleFileName = flag.String("f", "", "Module file name")
)

func main() {

	flag.Parse()

	file := loadFile()
	readSecrets(*file)

	err := file.Close()
	if err != nil {
		return
	}

}
