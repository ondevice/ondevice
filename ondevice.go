package main

import (
	"os"

	"github.com/ondevice/ondevice/command"
	"github.com/ondevice/ondevice/logg"
)

func main() {
	if len(os.Args) < 2 {
		logg.Fatal("Missing command! try 'ondevice help'")
	}

	//logg.Debug("-- args: ", os.Args[1:])
	cmd := os.Args[1]
	rc := command.Run(cmd, os.Args[2:])
	os.Exit(rc)
}
