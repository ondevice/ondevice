package command

import (
	"encoding/json"
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/filter"
	"github.com/ondevice/ondevice/logg"
)

type listCommand struct {
	BaseCommand
}

// ListOpts -- commandline arguments for `ondevice list`
var ListOpts struct {
	Properties bool   `short:"p" long:"props" description:"Include properties in the JSON output"`
	JSON       bool   `short:"j" long:"json" description:"JSON output, on object/device per line"`
	State      string `long:"state" description:"Filter output by device state, one of online/offline"`
	User       string `long:"user" description:"List devices of another user"`
	PrintIDs   bool   `long:"print-ids" description:"Print only "`
}

func (l listCommand) run(args []string) int {
	// parse args
	opts := ListOpts
	var filters []string
	var err error
	if filters, err = flags.ParseArgs(&opts, args); err != nil {
		logg.Fatal(err)
	}

	// check output flags
	if opts.JSON && opts.PrintIDs {
		logg.Fatal("Specified conflicting output modes (--json and --print-ids)")
	}

	// --user
	var auth api.Authentication
	if opts.User != "" {
		if auth, err = api.GetClientAuthForUser(opts.User); err != nil {
			logg.Fatalf("Can't find client auth for user '%s'", opts.User)
			return 1
		}
	} else {
		if auth, err = api.GetClientAuth(); err != nil {
			logg.Fatal("Missing client auth, have you run 'ondevice login'?")
			return 1
		}
	}

	// imply --props if filters have been specified
	if len(filters) > 0 {
		opts.Properties = true
	}

	allDevices, err := api.ListDevices(opts.State, opts.Properties, auth)
	if err != nil {
		logg.Fatal(err)
	}

	var devices = make([]api.Device, 0, len(allDevices))
	for _, dev := range allDevices {
		var ok bool
		if ok, err = _matches(dev, filters); err != nil {
			logg.Fatal(err)
		} else if ok {
			devices = append(devices, dev)
		}
	}

	if opts.JSON {
		l.printJSON(devices)
	} else if opts.PrintIDs {
		for _, dev := range devices {
			if _, err = fmt.Println(dev.ID); err != nil {
				logg.Error("print failed: ", err)
				break
			}
		}
	} else {
		l.print(devices)
	}

	return 0
}

func (l listCommand) print(devices []api.Device) {
	// find the maximum lengths for each column
	titles := []string{"ID", "State", "IP", "Version", "Name"}
	widths := []int{2, 5, 2, 7, 4}
	for _, dev := range devices {
		cols := _getColumns(dev)
		for j, col := range cols {
			width := len(col)
			if width > widths[j] {
				widths[j] = width
			}
		}
	}

	_printColumns(widths, titles)

	for _, dev := range devices {
		_printColumns(widths, _getColumns(dev))
	}
}

func (l listCommand) printJSON(devs []api.Device) {
	for _, dev := range devs {
		out, err := json.Marshal(dev)
		if err != nil {
			logg.Fatal("JSON serialization failed: ", err)
		}
		fmt.Println(string(out))
	}
}

func _getColumns(dev api.Device) []string {
	return []string{dev.ID, dev.State, dev.IP, dev.Version, dev.Name}
}

// returns true if the device matches the given --with(out) flags
func _matches(dev api.Device, filters []string) (bool, error) {
	for _, f := range filters {
		if ok, err := filter.Matches(dev, f); err != nil {
			return false, err
		} else if !ok {
			return false, nil
		}
	}

	return true, nil
}

func _printColumns(widths []int, cols []string) {
	if len(widths) != len(cols) {
		logg.Fatal("mismatch between cols and widths count", cols, widths)
	}

	for i, width := range widths {
		_printValue(width, cols[i])
		fmt.Print(" ")
	}
	fmt.Println("")
}

func _printValue(width int, val string) {
	if len(val) > width {
		logg.Fatal("width < len(val) !")
	}
	fmt.Print(val)
	for i := len(val); i < width; i++ {
		fmt.Print(" ")
	}
}

// ListCommand -- implements `ondevice list`
var ListCommand = listCommand{
	BaseCommand: BaseCommand{
		Arguments: "[--json] [--props] [--status=<online/offline>] [--print-ids] [filters...]",
		ShortHelp: "List your devices",
		RunFn:     nil, // we're implementing our own run() method
		LongHelp: `$ ondevice list [--options] [filters...]

List your devices

Options:
--json
  output JSON, one line/object per device
--props
  include properties (only affects JSON output)
--state <online/offline>
  limit to devices that are on/offline
--print-ids
	Print one devId per line (instead of tabular or JSON output)
	Can be used as input to

Filters:

With filters you can limit the output based on device properties.

Filters have the format:
  propertyName[operator[value]]
e.g.:
  somePackageVersion<2.3.4
  foo=

All comparisons are string-based (e.g. "234" is less than "99")
Missing properties are treated like the empty string ("arch=" lists devices with
empty or missing 'arch' property).


The following comparison operators are supported (comma-separated):
"=,==,!=,<,<<,<=,>,>>,>="

The one-character operators ("=,<,>") each have a two-character alias ("==,<<,>>"
respectively). These are provided to fix ambiguities (if the value to c start with "")
Note that you might have to escape '>' and '<' to shell redirection


Example:
  $ ondevice list
  ID            State   IP             Version         Name
  demo.7t91ta   offline                ondevice v0.4.3
  demo.fbqh2p   offline 192.168.1.23   ondevice v0.3.9
  demo.q5dkpm   online  127.0.0.1      ondevice v0.4.2
  demo.thm7br   offline 10.0.0.127     ondevice v0.4.3 My Raspberry PI

Example: Filters
  $ ondevice list 'fooVersion<2.3.4' 'foo=' --print-ids
	demo.7t91ta
  demo.fbqh2p
	demo.thm7br

	Only lists devices with the "fooVersion" property less than "2.3.4" (simple
	string comparison, so "2.3.4" < "2.34.5") and without the "foo" property
	(unset is equivalent to "")

Example: JSON output
  $ ondevice list --json --props
  {"id":"demo.7t91ta",state":"offline","stateTs":1490197318991,"version":"ondevice v0.4.3"}
  {"id":"demo.fbqh2p","ip":"192.168.1.23","state":"offline","stateTs":1485721709598,"version":"ondevice v0.3.9"}
  {"id":"demo.q5dkpm","ip":"127.0.0.1","state":"offline","stateTs":1487068641353,"version":"ondevice v0.4.2","props":{"test":"1234"}}
  {"id":"demo.thm7br","ip":"10.0.0.127","state":"offline", "": "My Raspberry PI","stateTs":1490963689912,"version":"ondevice v0.4.3"}

  Note that JSON fields may be missing or null

`,
	},
}
