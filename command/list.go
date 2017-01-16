package command

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jessevdk/go-flags"
	"github.com/ondevice/ondevice-cli/rest"
)

// ListCmd -- `ondevice list` implementation
type ListCmd struct{}

// ListOpts -- commandline arguments for `ondevice list`
var ListOpts struct {
	Properties bool   `short:"p" long:"props" description:"Include properties in the JSON output"`
	JSON       bool   `short:"j" long:"json" description:"JSON output, on object/device per line"`
	State      string `long:"state" description:"Filter output by device state, one of online/offline"`
}

const _longListHelp = `ondevice list

List your devices

Options:
--json
  output JSON, one line/object per device
--props
  include properties (only affects JSON output)
--state <online/offline>
  limit to devices that are on/offline
`

func (l ListCmd) args() []string {
	return nil
}

func (l ListCmd) longHelp() string {
	return _longListHelp
}

func (l ListCmd) shortHelp() string {
	return "List your devices"
}

// Run -- implements `ondevice list`
func (l ListCmd) Run(args []string) {
	// parse args
	opts := ListOpts
	if _, err := flags.ParseArgs(&opts, args); err != nil {
		log.Fatal(err)
	}

	devices, err := rest.ListDevices(opts.State, opts.Properties)
	if err != nil {
		log.Fatal(err)
	}

	if opts.JSON {
		l.printJSON(devices)
	} else {
		l.print(devices)
	}
}

func (l ListCmd) print(devices []rest.Device) {
	// find the maximum lengths for each column
	titles := []string{"ID", "State", "IP", "Version", "Name"}
	widths := []int{2, 5, 2, 7, 4}
	for i := range devices {
		dev := devices[i]
		cols := _getColumns(dev)
		for j := range cols {
			width := len(cols[j])
			if width > widths[j] {
				widths[j] = width
			}
		}
	}

	_printColumns(widths, titles)

	for i := range devices {
		dev := devices[i]
		_printColumns(widths, _getColumns(dev))
	}
}

func (l ListCmd) printJSON(devs []rest.Device) {
	for i := range devs {
		dev := devs[i]
		out, err := json.Marshal(dev)
		if err != nil {
			log.Fatal("JSON serialization failed: ", err)
		}
		fmt.Println(string(out))
	}
}

func _getColumns(dev rest.Device) []string {
	return []string{dev.ID, dev.State, dev.IP, dev.Version, dev.Name}
}

func _printColumns(widths []int, cols []string) {
	if len(widths) != len(cols) {
		log.Fatal("mismatch between cols and widths count", cols, widths)
	}

	for i := range widths {
		_printValue(widths[i], cols[i])
		fmt.Print(" ")
	}
	fmt.Println("")
}

func _printValue(width int, val string) {
	if len(val) > width {
		log.Fatal("width < len(val) !")
	}
	fmt.Print(val)
	for i := len(val); i < width; i++ {
		fmt.Print(" ")
	}
}
