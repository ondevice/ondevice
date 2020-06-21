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
	"os"

	"github.com/ondevice/ondevice/cmd/internal"
	"github.com/ondevice/ondevice/config"
	"github.com/spf13/cobra"
)

// sshCmd represents the ssh command
type sshCmd struct {
	cobra.Command
}

func init() {
	var c sshCmd

	c.Command = cobra.Command{
		Use:   "ssh [ssh-arguments...]",
		Short: "connect to your devices using the ssh protocol",
		Long: `Connect to your devices using the 'ssh' command.

This is a relatively thin wrapper around the 'ssh' command.
The main difference to invoking ssh directly is that instead of regular host names you'll have to specify an ondevice deviceId.
The connection is routed through the ondevice.io network.

ondevice ssh will try to parse ssh's arguments, the first non-argument has to be
the user@devId combo.

See ssh's documentation for further details.

Notes:
- We use our own known_hosts file (in ~/.config/ondevice/known_hosts).
  Override with ''-oUserKnownHostsFile=...'`,
		Example: `- simply connect to device1:
  $ ondevice ssh device1

- open an SSH connection to device1, logging in as 'user':
  $ ondevice ssh user@device1

- run 'echo hello world' on device1:
  $ ondevice ssh device1 echo hello world

- tunnel the HTTP server on device1 to the local port 1234 without opening
  a shell:
  $ ondevice ssh device1 -N -L 1234:localhost:80

- start a SOCKS5 proxy listening on port 1080. It'll redirect all traffic
  to the target host:
  $ ondevice ssh device1 -D 1080`,
		Run:               c.run,
		ValidArgsFunction: internal.DeviceListCompletion{}.Run,
	}
	c.DisableFlagParsing = true

	rootCmd.AddCommand(&c.Command)
}

func (c *sshCmd) run(cmd *cobra.Command, args []string) {
	var sshCommand = config.MustLoad().GetCommand(config.CommandSSH)

	// we use the ProxyCommand option to have ssh invoke 'ondevice pipe %h ssh'
	sshCommand = append(sshCommand, fmt.Sprintf("-oProxyCommand='%s' pipe %%h ssh", os.Args[0]))

	// use our own known_hosts file
	if knownHostsFile := config.MustLoad().GetFilePath(config.PathKnownHosts); knownHostsFile != "" {
		sshCommand = append(sshCommand, fmt.Sprintf("-oUserKnownHostsFile=%s", knownHostsFile))
	}

	sshCommand = append(sshCommand, args...)

	// ExecExternalCommand won't return
	internal.ExecExternalCommand(sshCommand[0], sshCommand)
}
