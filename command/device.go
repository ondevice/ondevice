package command

import (
	"fmt"
	"log"
	"strings"

	"github.com/ondevice/ondevice-cli/rest"
)

// DeviceCmd -- `ondevice device` implementation
type DeviceCmd struct{}

func (d DeviceCmd) args() string {
	return "<devId> <props/set/rm> [...]"
}

func (d DeviceCmd) longHelp() string {
	log.Fatal("impl me")
	return ""
}

func (d DeviceCmd) shortHelp() string {
	return "List/manipulate device properties"
}

// Run -- `ondevice device` implementation
func (d DeviceCmd) Run(args []string) int {
	if len(args) < 1 {
		log.Fatal("Error: missing deviceId")
	} else if len(args) < 2 {
		log.Fatal("Error: missing device command")
	} else if args[1] == "set" {
		d.setProperties(args[0], args[2:])
	} else if args[1] == "rm" {
		d.removeProperties(args[0], args[2:])
	} else if args[1] == "props" || args[1] == "properties" || args[1] == "list" {
		d.listProperties(args[0])
	} else {
		log.Fatal("Unknown device command: ", args[1])
	}

	return 0
}

func (d DeviceCmd) listProperties(devID string) {
	_printProperties(rest.ListProperties(devID))
}

func (d DeviceCmd) removeProperties(devID string, args []string) {
	_printProperties(rest.RemoveProperties(devID, args))
}

func (d DeviceCmd) setProperties(devID string, args []string) {
	var props = make(map[string]string)

	for i := range args {
		s := strings.SplitN(args[i], "=", 2)
		if _, ok := props[s[0]]; ok {
			log.Fatalf("Duplicate value for property '%s'", s[0])
		}
		props[s[0]] = s[1]
	}

	_printProperties(rest.SetProperties(devID, props))
}

func _printProperties(props map[string]string, err error) {
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range props {
		fmt.Printf("%s=%s\n", k, v)
	}
}
