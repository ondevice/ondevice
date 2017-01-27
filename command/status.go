package command

import (
	"encoding/json"
	"fmt"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/control"
	"github.com/ondevice/ondevice/logg"
)

// StatusCmd -- implements 'ondevice status'
type StatusCmd struct{}

const _longStatusHelp = `ondevice status [--json]

Print status and version information on the local ondevice client/device

Options:
--json
  Prints JSON formatted instead of human readable output

Exit Codes:
0: daemon up and running
1: on error
2: daemon running but not online
3: daemon not running (or unreachable)
`

func (s *StatusCmd) args() string {
	return "[--json]"
}

func (s *StatusCmd) longHelp() string {
	return _longStatusHelp
}

func (s *StatusCmd) shortHelp() string {
	return "Prints the client and local device status"
}

// Run -- implements 'ondevice status'
func (s *StatusCmd) Run(args []string) int {
	if len(args) == 0 {
		return print()
	} else if len(args) != 1 || args[0] != "--json" {
		logg.Fatal("'ondevice status' expects either no arguments or a single '--json'")
		return 1
	} else {
		return printJSON()
	}
}

func print() int {
	state := getState()

	if state.Device != nil {
		fmt.Println("Device:")
		fmt.Println("  devID: ", state.Device["devId"])
		fmt.Println("  state: ", state.Device["state"])
		fmt.Println("  version: ", state.Version)
		fmt.Println("")
	}

	fmt.Println("Client:")
	fmt.Println("  version: ", state.Client["version"])

	return getReturnCode(state)
}

func printJSON() int {
	state := getState()

	buff, _ := json.MarshalIndent(state, "", "  ")
	fmt.Println(string(buff))

	return getReturnCode(state)
}

func getReturnCode(state control.DeviceState) int {
	if state.Device == nil {
		return 3 // missing "device" status -> daemon not running
	}

	if ds, ok := state.Device["state"]; ok {
		if ds == "online" {
			return 0
		}
	}
	return 1
}

func getState() control.DeviceState {
	rc, err := control.GetState()
	if err != nil {
		// can't query device socket -> assume daemon is not running
		// TODO make me a debug message
		logg.Warning("Couldn't query device state: ", err)
		rc = control.DeviceState{}
	}

	rc.Client = map[string]string{
		"version": config.GetVersion(),
	}

	return rc
}
