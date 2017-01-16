package command

import (
	"fmt"
	"log"

	"github.com/ondevice/ondevice-cli/rest"
)

// DeviceCmd -- `ondevice device` implementation
type DeviceCmd struct{}

func (d DeviceCmd) args() []string {
	return nil
}

func (d DeviceCmd) longHelp() string {
	log.Fatal("impl me")
	return ""
}

func (d DeviceCmd) shortHelp() string {
	return "List/manipulate device properties"
}

// Run -- `ondevice device` implementation
func (d DeviceCmd) Run(args []string) {
	if len(args) < 1 {
		log.Fatal("Error: missing deviceId")
	} else if len(args) < 2 {
		log.Fatal("Error: missing device command")
	} else if args[1] == "props" || args[1] == "properties" {
		d.listProperties(args[0])
	} else {
		log.Fatal("Unknown device command: ", args[1])
	}
}

func (d DeviceCmd) listProperties(devID string) {
	props := rest.ListProperties(devID)

	for k, v := range props {
		fmt.Printf("%s=%s\n", k, v)
	}
}
