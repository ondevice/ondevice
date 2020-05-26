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
	"os"
	"syscall"
	"time"

	"github.com/ondevice/ondevice/daemon"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stops the local ondevice daemon (if running)",
	Long: `Stops a running ondevice daemon (using the ondevice.pid file) and tries to
terminate it.

Returns 0 if the daemon was stopped, 1 if it wasn't running and 2 on error
(e.g. after a 30sec timeout)`,
	Run: stopRun,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func stopRun(cmd *cobra.Command, args []string) {
	var found = false
	for i := 0; i < 30; i++ {
		// fetch the PID every time (the daemon might go inactive - still running until the remaining tunnels close)
		p, err := daemon.GetDaemonProcess()
		if err != nil {
			if !found {
				logrus.WithError(err).Error("couldn't find daemon process")
				os.Exit(1)
			}
			// we seem to have stopped it
			os.Exit(1)
		}

		if !found {
			logrus.Infof("stopping ondevice daemon... (pid: %d)", p.Pid)
			found = true
		}

		p.Signal(syscall.SIGTERM)
		time.Sleep(1000 * time.Millisecond)
	}

	logrus.Error("timeout trying to stop the ondevice daemon")
	os.Exit(2)
}
