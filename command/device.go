package command

import (
	"fmt"
	"strings"

	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/logg"
)

// DeviceCmd -- `ondevice device` implementation
type DeviceCmd struct{}

func (d DeviceCmd) args() string {
	return "<devId> <props/set/rm> [...]"
}

func (d DeviceCmd) longHelp() string {
	logg.Fatal("impl me")
	return ""
}

func (d DeviceCmd) shortHelp() string {
	return "List/manipulate device properties"
}

// Run -- `ondevice device` implementation
func (d DeviceCmd) Run(args []string) int {
	if len(args) < 1 {
		logg.Fatal("Error: missing deviceId")
	} else if len(args) < 2 {
		logg.Fatal("Error: missing device command")
	} else if args[1] == "set" {
		d.setProperties(args[0], args[2:])
	} else if args[1] == "rm" {
		d.removeProperties(args[0], args[2:])
	} else if args[1] == "props" || args[1] == "properties" || args[1] == "list" {
		d.listProperties(args[0])
	} else {
		logg.Fatal("Unknown device command: ", args[1])
	}

	return 0
}

func (d DeviceCmd) listProperties(devID string) {
	_printProperties(api.ListProperties(devID))
}

func (d DeviceCmd) removeProperties(devID string, args []string) {
	_printProperties(api.RemoveProperties(devID, args))
}

func (d DeviceCmd) setProperties(devID string, args []string) {
	var props = make(map[string]string)

	for i := range args {
		s := strings.SplitN(args[i], "=", 2)
		if _, ok := props[s[0]]; ok {
			logg.Fatalf("Duplicate value for property '%s'", s[0])
		}
		props[s[0]] = s[1]
	}

	_printProperties(api.SetProperties(devID, props))
}

func _printProperties(props map[string]string, err error) {
	if err != nil {
		logg.Fatal(err)
	}

	for k, v := range props {
		fmt.Printf("%s=%s\n", k, v)
	}
}
