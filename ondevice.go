package main

import (
	"log"
	"os"

	"github.com/ondevice/ondevice-cli/command"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing command! try 'ondevice help'")
	}

	//log.Print("-- args: ", os.Args[1:])
	cmd := os.Args[1]
	command.Run(cmd, os.Args[2:])
}
