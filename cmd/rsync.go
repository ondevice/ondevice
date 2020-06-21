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

// rsyncCmd represents the rsync command
var rsyncCmd = &cobra.Command{
	Use:   "rsync [rsync args...]",
	Short: "copy files from/to your devices using rsync",
	Long: `copy files from/to devices using rsync

This command is only a thin wrapper around the 'rsync' client (using its '-e'
argument to make it use 'ondevice ssh' internally).

Have a look at the rsync man page for further details.`,
	Example: `- copy the local /src/path to myDev's /target/path/ as root
  (and pass the -a and -v options to rsync)
  $ ondevice rsync -av /source/path/ root@myDev:/target/path/

- copy otherDev's /etc/motd file to /tmp/other.motd (and login as 'me')
  $ ondevice rsync me@otherDev:/etc/motd /tmp/other.motd`,
	Run:               rsyncRun,
	ValidArgsFunction: scpValidate,
}

func init() {
	rootCmd.AddCommand(rsyncCmd)
	rsyncCmd.DisableFlagParsing = true
}

func rsyncRun(cmd *cobra.Command, args []string) {
	var rsyncCommand = config.MustLoad().GetCommand(config.CommandRSYNC).Value()

	// TODO this will fail if argv[0] contains spaces
	rsyncCommand = append(rsyncCommand, "-e", fmt.Sprintf("'%s' ssh", os.Args[0]))
	rsyncCommand = append(rsyncCommand, args...)

	// ExecExternalCommand won't return (potential errors will cause logrus.Fatal() calls)
	internal.ExecExternalCommand(rsyncCommand[0], rsyncCommand)
}
