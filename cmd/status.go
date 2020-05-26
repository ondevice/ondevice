/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/control"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "prints the client and local device status",
	Long: `Print status and version information on the local ondevice client/device

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
	  }`,
	Run: statusRun,
}

var jsonFlag bool

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVar(&jsonFlag, "json", false, "prints JSON formatted instead of human readable output")
}

func statusRun(cmd *cobra.Command, args []string) {
	var rc int
	if len(args) != 0 {
		logrus.Fatal("too many arguments")
	}

	if jsonFlag {
		rc = statusPrintJSON()
	} else {
		rc = statusPrint()
	}

	if rc != 0 {
		os.Exit(rc)
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
		logrus.WithError(err).Debug("couldn't query device state")
		rc = control.DeviceState{}
	}

	rc.Client = map[string]string{
		"version": config.GetVersion(),
	}

	return rc
}
