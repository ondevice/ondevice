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

func statusRun(args []string) int {
	if len(args) == 0 {
		return statusPrint()
	} else if len(args) != 1 || args[0] != "--json" {
		logg.Fatal("'ondevice status' expects either no arguments or a single '--json'")
		return 1
	} else {
		return statusPrintJSON()
	}
}

func statusPrint() int {
	state := statusGetState()

	if state.Device != nil {
		fmt.Println("Device:")
		fmt.Println("  devID: ", state.Device["devId"])
		fmt.Println("  state: ", state.Device["state"])
		fmt.Println("  version: ", state.Version)
		fmt.Println("")
	}

	fmt.Println("Client:")
	fmt.Println("  version: ", state.Client["version"])

	return statusGetReturnCode(state)
}

func statusPrintJSON() int {
	state := statusGetState()

	buff, _ := json.MarshalIndent(state, "", "  ")
	fmt.Println(string(buff))

	return statusGetReturnCode(state)
}

func statusGetReturnCode(state control.DeviceState) int {
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

func statusGetState() control.DeviceState {
	rc, err := control.GetState()
	if err != nil {
		// can't query device socket -> assume daemon is not running
		logg.Debug("Couldn't query device state: ", err)
		rc = control.DeviceState{}
	}

	rc.Client = map[string]string{
		"version": config.GetVersion(),
	}

	return rc
}

// StatusCommand -- implements `ondevice status`
var StatusCommand = BaseCommand{
	Arguments: "[--json]",
	ShortHelp: "Prints the client and local device status",
	LongHelp: `$ ondevice status [--json]

Print status and version information on the local ondevice client/device

Options:
--json
  Prints JSON formatted instead of human readable output

Exit Codes:
0: daemon up and running
1: on error
2: daemon running but not online
3: daemon not running (or unreachable)

Examples:

  $ ondevice status
  Device:
    devID:  demo.q5dkpm
    state:  online
    version:  0.4.4

  Client:
    version:  0.4.4

  $ ondevice status --json
  {
    "version": "0.4.4",
    "client": {
      "version": "0.4.4"
    },
    "device": {
      "devId": "demo.q5dkpm",
      "state": "online"
    }
  }
`,
	RunFn: statusRun,
}
