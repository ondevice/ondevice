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
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/control"
	"github.com/ondevice/ondevice/daemon"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the ondevice device daemon",
	Long: `Starts the ondevice daemon (the device side of the tunnels).

On debian based systems instead of running 'ondevice daemon' directly, you
should install the ondevice-daemon package instead (which will also take care
of setting up the credentials)

Make sure you run 'ondevice login' and authenticate with your device key first.

Usually you'll only want to have one 'ondevice daemon' instance per device.
If you want to run multiple, you'll have to specify the .conf, .pid and .sock
files manually.
Concurrent daemon instances can't use the same ondevice.conf file!!!

Example Socket URLs:
- unix:///home/user/.config/ondevice/ondevice.sock
  User's ondevice.sock path - clients will use this URL first when connecting
- unix:///var/run/ondevice/ondevice.sock
  Default system-wide ondevice.sock path - if the above failed, clients will try
  this one instead.
- /var/run/ondevice/ondevice.sock
  Same as the above (since unix:// is the default URL scheme here)
- http://localhost:1234/
	Listen on TCP port 1234.
  Note that there's currently support for neither SSL nor authentication so use
  this only if absolutely necessary

On the client side, set the ONDEVICE_HOST environment variable to match the
socket parameter.
`,
	Run: daemonRun,
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().String("pidfile", "", "if set, overrides the config value set in path.ondevice_pid")
	daemonCmd.Flags().String("sock", "", "if set, overrides the config vlaue set in path.ondevice_sock")
}

func daemonRun(cmd *cobra.Command, args []string) {
	// main.go disables timestamps in log messages. re-enable them
	log.SetFlags(log.LstdFlags)

	if os.Getuid() == 0 {
		logrus.Fatal("`ondevice daemon` should not be run as root")
	}

	var d = daemon.NewDaemon()
	var controlURL *url.URL
	var err error

	if controlURL, err = daemonParseArgs(cmd, args, d); err != nil {
		logrus.WithError(err).Fatal("failed to parse daemon args")
		return
	}

	c := control.NewSocket(d, *controlURL)
	d.Control = c

	d.Run()
}

// Parses the commandline arguments, returns the ControlSocket URL
func daemonParseArgs(cmd *cobra.Command, args []string, d *daemon.Daemon) (*url.URL, error) {
	var err error

	if len(args) > 0 {
		return nil, fmt.Errorf("Too many arguments: %s", args)
	}

	var cfg = config.MustLoad()

	// override config values if flag is set
	if pidFlag := cmd.Flag("pidfile").Value.String(); pidFlag != "" {
		if err = cfg.SetValue(config.PathOndevicePID, pidFlag); err != nil {
			logrus.WithError(err).Error("failed to override daemon PID file")
			return nil, err
		}
	}
	if sockURL := cmd.Flag("sock").Value.String(); sockURL != "" {
		if err = cfg.SetValue(config.PathOndeviceSock, sockURL); err != nil {
			logrus.WithError(err).Error("failed to override daemon socket")
			return nil, err
		}
	}

	var pidFile = cfg.GetPath(config.PathOndevicePID)
	if pidFile.Error() != nil {
		logrus.WithError(err).Error("failed to get daemon PID file")
		return nil, err
	}
	d.PIDFile = pidFile.GetAbsolutePath()

	var sockURL = cfg.GetPath(config.PathOndeviceSock)
	if sockURL.Error() != nil {
		logrus.WithError(err).Error("failed to get daemon socket URL")
		return nil, err
	}

	return sockURL.GetAbsoluteURL(), nil
}
