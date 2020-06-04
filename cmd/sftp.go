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
	"strings"
	"syscall"

	"github.com/ondevice/ondevice/cmd/internal"
	"github.com/ondevice/ondevice/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var sftpFlags = sshParseFlags("1246aCfpqrvB:b:c:D:F:i:l:o:P:R:S:s:")

// sftpCmd represents the sftp command
var sftpCmd = &cobra.Command{
	Use:   "sftp [sftp-flags] [user@]devId",
	Short: "copy files from/to a device using sftp",
	Long: `interactively copy files from/to devices using sftp

Notes:
- We use our own known_hosts file (in ~/.config/ondevice/known_hosts).
  Override with ''-oUserKnownHostsFile=...'`,
	Example: `- open an sftp session to 'myDev', logging in as 'user'
  $ ondevice sftp user@myDev`,
	Run:               sftpRun,
	ValidArgsFunction: internal.DeviceListCompletion{}.Run,
}

func init() {
	rootCmd.AddCommand(sftpCmd)
	sftpCmd.DisableFlagParsing = true
}

func sftpRun(cmd *cobra.Command, args []string) {
	sftpPath := "/usr/bin/sftp"

	args, opts := sshParseArgs(sftpFlags, args)

	// parse all the args as possible remote files [[user@]devId]:/path/to/file
	for i := 0; i < len(args); i++ {
		var arg = args[i]
		if strings.HasPrefix(arg, "./") || strings.HasPrefix(arg, "/") {
			// absolute/relative filename -> local file (this if block allows copying of files with colons in them)
			continue
		} else if parts := strings.SplitN(arg, ":", 2); len(parts) == 2 {
			// remote file -> parse and transform user@host part
			// results in "[user@]account.devId"
			var tgtHost, tgtUser = sshParseTarget(parts[0])
			if tgtUser != "" {
				tgtHost = fmt.Sprintf("%s@%s", tgtUser, tgtHost)
			}
			args[i] = fmt.Sprintf("%s:%s", tgtHost, parts[1])
		}
	}

	// TODO this will fail if argv[0] contains spaces
	a := []string{sftpPath, fmt.Sprintf("-oProxyCommand=%s pipe %%h ssh", os.Args[0])}
	if sshGetConfig(opts, "UserKnownHostsFile") == "" {
		var cfg, err = config.Read()
		if err != nil {
			logrus.WithError(err).Fatal("failed to read config")
			return
		}
		a = append(a, fmt.Sprintf("-oUserKnownHostsFile=%s", cfg.GetFilePath("known_hosts")))
	}

	a = append(a, opts...)
	a = append(a, args...)

	err := syscall.Exec(sftpPath, a, os.Environ())
	if err != nil {
		logrus.WithError(err).Fatalf("failed to run '%s'", sftpPath)
	}

	logrus.Fatal("we shouldn't be here")
}
