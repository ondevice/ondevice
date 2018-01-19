package command

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/logg"
)

type DeviceOpts struct {
	Yes bool `long:"yes" short:"y" description:"Confirm deletion noninteractively"`
}

func deviceRun(args []string) int {
	var opts DeviceOpts
	var err error

	if args, err = flags.ParseArgs(&opts, args); err != nil {
		logg.Fatal(err)
	}

	if len(args) < 1 {
		err = errors.New("missing deviceId")
	} else if len(args) < 2 {
		err = errors.New("missing device command")
	} else if args[1] == "set" {
		err = deviceSetProperties(args[0], args[2:])
	} else if args[1] == "rm" {
		err = deviceRemoveProperties(args[0], args[2:], opts)
	} else if args[1] == "props" || args[1] == "properties" || args[1] == "list" {
		err = deviceListProperties(args[0])
	} else {
		err = fmt.Errorf("Unknown device command: '%s'", args[1])
	}

	if err != nil {
		logg.Fatal("Error: ", err)
		return 1
	}
	return 0
}

func deviceListProperties(devID string) error {
	return _printProperties(api.ListProperties(devID))
}

func deviceRemoveProperties(devID string, args []string, opts DeviceOpts) error {
	if len(args) == 0 {
		logg.Error("Too few arguments")
	}

	// check if the user wants to delete the device ("on:id" present)
	var wantsDelete = false
	for _, key := range args {
		if key == "on:id" {
			wantsDelete = true
			break
		}
	}

	if wantsDelete {
		if len(args) != 1 {
			logg.Fatal("To delete a device, remove its 'on:id' property (and nothing else)")
		}

		var confirmed = opts.Yes
		if !confirmed {
			var reader = bufio.NewReader(os.Stdin)
			var input string
			var err error

			for input == "" {
				fmt.Printf("Do you really want to delete the device '%s' (y/N): ", devID)
				input, err = reader.ReadString('\n')
				if err != nil {
					logg.Fatal(err)
				}

				switch strings.TrimSpace(strings.ToLower(input)) {
				case "y", "yes":
					confirmed = true
				case "n", "no", "":
					confirmed = false
					input = "no"
				default:
					input = ""
				}
			}
		}

		if confirmed {
			if err := api.DeleteDevice(devID); err != nil {
				return err
			}
		} else {
			return errors.New("Aborted delete")
		}
		return nil
	}

	return _printProperties(api.RemoveProperties(devID, args))
}

func deviceSetProperties(devID string, args []string) error {
	var props = make(map[string]string)

	for _, arg := range args {
		s := strings.SplitN(arg, "=", 2)
		if _, ok := props[s[0]]; ok {
			return fmt.Errorf("Duplicate value for property '%s'", s[0])
		}
		props[s[0]] = s[1]
	}

	return _printProperties(api.SetProperties(devID, props))
}

func _printProperties(props map[string]interface{}, err error) error {
	if err != nil {
		return err
	}

	// get list of keys and sort them
	var keys = make([]string, 0, len(props))
	for k := range props {
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

	return nil
}

// DeviceCommand -- implemnts `ondevice device`
var DeviceCommand = BaseCommand{
	Arguments: "<devId> <props/set/rm> [key1=val1 ...]",
	ShortHelp: "List/manipulate device properties",
	RunFn:     deviceRun,
	LongHelp: `$ ondevice device <devId> props
$ ondevice device <devId> set [key1=val1 ...]
$ ondevice device <devId> rm [--yes/-y] [--delete] [key1 key2 ...]

This command allows you to change all your devices' properties.
It requires a client key with the 'manage' authorization.

Properties can be used to keep track of your devices, to manage their characteristics,
keep tracks of running maintenance scripts, etc.

- ondevice device $devId props
  lists that device's properties, one per line, as 'key=value' pairs
- ondevice device $devId set [key=val...]
  sets one or more device properties, again as 'key=value' pairs
- ondevice device $devId rm [key ...]
  removes one or more device properties by name

Some special cases are:
- ondevice device $devId set on:id=newId
  Rename (= change devId of) a device
- ondevice device $devId rm on:id
  Removing the special property 'on:id' will attempt to delete the device
  (will ask for confirmation unless you also specify --yes)
  Only devices that have been offline for at least an hour can be deleted.

Options:
--yes -y
  Don't ask before deleting a device

Each invocation will print the resulting property list.

Examples:
  $ ondevice device q5dkpm props
  $ ondevice device q5dkpm set test=1234 foo=bar
  test=1234
  foo=bar
  $ ondevice device q5dkpm rm foo
  test=1234

  # rename and then delete the device
  $ ondevice device q5dkpm set on:id=rpi
  $ ondevice device rpi rm on:id
  Do you really want to delete the device 'uyqsn4' (y/N):
`,
}
