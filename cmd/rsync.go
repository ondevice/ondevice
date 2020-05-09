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
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rsyncCmd represents the rsync command
var rsyncCmd = &cobra.Command{
	Use:   "rsync [rsync args...]",
	Short: "copy files from/to your devices using rsync",
	Long: `copy files from/to devices using rsync

	Examples:

	- ondevice rsync -av /source/path/ root@myDev:/target/path/
	    copy the local /src/path to myDev's /target/path/ as root
	    (and pass the -a and -v options to rsync)
	- ondevice rsync me@otherDev:/etc/motd /tmp/other.motd
	    copy otherDev's /etc/motd file to /tmp/other.motd (and login as 'me')

	This command is only a thin wrapper around the 'rsync' client (using its '-e'
	argument to make it use 'ondevice ssh' internally).

	Have a look at the rsync man page for further details.
`,
	Run: rsyncRun,
}

func init() {
	rootCmd.AddCommand(rsyncCmd)
	rsyncCmd.DisableFlagParsing = true
}

func rsyncRun(cmd *cobra.Command, args []string) {
	rsyncPath := "/usr/bin/rsync"

	// TODO this will fail if argv[0] contains spaces
	a := []string{rsyncPath, "-e", fmt.Sprintf("%s ssh", os.Args[0])}
	a = append(a, args...)

	err := syscall.Exec(rsyncPath, a, os.Environ())
	if err != nil {
		logrus.WithError(err).Fatalf("failed to run '%s'", rsyncPath)
	}

	logrus.Fatal("we shouldn't be here")
}
