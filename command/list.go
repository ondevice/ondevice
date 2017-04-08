package command

import (
	"encoding/json"
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/logg"
)

// ListCmd -- `ondevice list` implementation
type ListCmd struct{}

// ListOpts -- commandline arguments for `ondevice list`
var ListOpts struct {
	Properties bool   `short:"p" long:"props" description:"Include properties in the JSON output"`
	JSON       bool   `short:"j" long:"json" description:"JSON output, on object/device per line"`
	State      string `long:"state" description:"Filter output by device state, one of online/offline"`
}

const _longListHelp = `
ondevice list

List your devices

Options:
--json
  output JSON, one line/object per device
--props
  include properties (only affects JSON output)
--state <online/offline>
  limit to devices that are on/offline

Example:

  $ ondevice list
  ID                State   IP             Version         Name
  demo.7t91ta   offline                ondevice v0.4.3
  demo.fbqh2p   offline 192.168.1.23   ondevice v0.3.9
  demo.q5dkpm   online  127.0.0.1      ondevice v0.4.2
  demo.thm7br   offline 10.0.0.127     ondevice v0.4.3

  $ ondevice list --json --props
  {"id":"demo.7t91ta",state":"offline","stateTs":1490197318991,"version":"ondevice v0.4.3"}
  {"id":"demo.fbqh2p","ip":"192.168.1.23","state":"offline","stateTs":1485721709598,"version":"ondevice v0.3.9"}
  {"id":"demo.q5dkpm","ip":"127.0.0.1","state":"offline","stateTs":1487068641353,"version":"ondevice v0.4.2","props":{"test":"1234"}}
  {"id":"demo.thm7br","ip":"10.0.0.127","state":"offline","stateTs":1490963689912,"version":"ondevice v0.4.3"}

Keep in mind that JSON fields may be missing or null

`

func (l ListCmd) args() string {
	return "[--json] [--props] [--status=<online/offline>]"
}

func (l ListCmd) longHelp() string {
	return _longListHelp
}

func (l ListCmd) shortHelp() string {
	return "List your devices"
}

func (l ListCmd) run(args []string) int {
	// parse args
	opts := ListOpts
	if _, err := flags.ParseArgs(&opts, args); err != nil {
		logg.Fatal(err)
	}

	devices, err := api.ListDevices(opts.State, opts.Properties)
	if err != nil {
		logg.Fatal(err)
	}

	if opts.JSON {
		l.printJSON(devices)
	} else {
		l.print(devices)
	}

	return 0
}

func (l ListCmd) print(devices []api.Device) {
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

func (l ListCmd) printJSON(devs []api.Device) {
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
