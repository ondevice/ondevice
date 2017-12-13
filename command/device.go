package command

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/logg"
)

func deviceRun(args []string) int {
	if len(args) < 1 {
		logg.Fatal("Error: missing deviceId")
	} else if len(args) < 2 {
		logg.Fatal("Error: missing device command")
	} else if args[1] == "set" {
		deviceSetProperties(args[0], args[2:])
	} else if args[1] == "rm" {
		deviceRemoveProperties(args[0], args[2:])
	} else if args[1] == "props" || args[1] == "properties" || args[1] == "list" {
		deviceListProperties(args[0])
	} else {
		logg.Fatal("Unknown device command: ", args[1])
	}

	return 0
}

func deviceListProperties(devID string) {
	_printProperties(api.ListProperties(devID))
}

func deviceRemoveProperties(devID string, args []string) {
	_printProperties(api.RemoveProperties(devID, args))
}

func deviceSetProperties(devID string, args []string) {
	var props = make(map[string]string)

	for _, arg := range args {
		s := strings.SplitN(arg, "=", 2)
		if _, ok := props[s[0]]; ok {
			logg.Fatalf("Duplicate value for property '%s'", s[0])
		}
		props[s[0]] = s[1]
	}

	_printProperties(api.SetProperties(devID, props))
}

func _printProperties(props map[string]interface{}, err error) {
	if err != nil {
		logg.Fatal(err)
	}

	// get list of keys and sort them
	var keys = make([]string, 0, len(props))
	for k, _ := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		var v = props[k]
		var repr string

		if s, ok := v.(string); ok {
			repr = s
		} else {
			var reprBytes, _ = json.Marshal(v)
			repr = string(reprBytes)
		}

		fmt.Printf("%s=%s\n", k, repr)
	}
}

// DeviceCommand -- implemnts `ondevice device`
var DeviceCommand = BaseCommand{
	Arguments: "<devId> <props/set/rm> [key1=val1 ...]",
	ShortHelp: "List/manipulate device properties",
	RunFn:     deviceRun,
	LongHelp: `$ ondevice device <devId> props
$ ondevice device <devId> set [key1=val1 ...]
$ ondevice device <devId> rm [key1 key2 ...]

This command allows you to change all your devices' properties.
It requires a client key with the 'manage' authorization.

Properties can be used to keep track of your devices, to manage their characteristics,
keep tracks of running maintenance scripts, etc.

- ondevice device <devId> props
  lists that device's properties, one per line, as 'key=value' pairs
- ondevice device <devId> set [key=val...]
  sets one or more device properties, again as 'key=value' pairs
- ondevice device <devId> rm [key ...]
  removes one or more device properties by name

Each of the invocations will print the resulting property list.

Example:
  $ ondevice device q5dkpm props
  $ ondevice device q5dkpm set test=1234 foo=bar
  test=1234
  foo=bar
  $ ondevice device q5dkpm rm foo
  test=1234
`,
}
