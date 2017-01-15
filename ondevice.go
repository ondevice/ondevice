package main

import "github.com/ondevice/ondevice-cli/command"

import "os"

func main() {
	cmd := os.Args[1]

	command.Run(cmd, os.Args[2:])
}
