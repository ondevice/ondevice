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

	"github.com/ondevice/ondevice/cmd/internal"
	"github.com/ondevice/ondevice/config"
	"github.com/spf13/cobra"
)

// scpCmd represents the scp command
var scpCmd = &cobra.Command{
	Use:   "scp [scp args...] [[user1@]host1:]file1 ... [[userN@]hostN:]fileN",
	Short: "copy files from/to your devices using scp",
	Long: `copy files from/to devices using scp

Notes:
- while it is possible to copy files between two servers, scp will initiate
  both connections simultaneously causing two concurrent password prompts
  which won't work (unless of course you've set up ssh_agent properly).
- uses scp's '-3' flag (allowing files to be copied between two remote devices)
- We use our own known_hosts file (in ~/.config/ondevice/known_hosts).
  Override with ''-oUserKnownHostsFile=...'`,
	Example: `- copy the local /src/path to myDev's /target/path/ as root
  (and pass the -p and -v options to scp)
  $ ondevice scp -pv /source/path/ root@myDev:/target/path/

- copy otherDev's /etc/motd file to /tmp/other.motd (and login as 'me')
  $ ondevice scp me@otherDev:/etc/motd /tmp/other.motd`,
	Run:               scpRun,
	ValidArgsFunction: scpValidate,
}

func init() {
	rootCmd.AddCommand(scpCmd)
	scpCmd.DisableFlagParsing = true
}

var scpFlags = sshParseFlags("12346BCpqrvc:F:i:l:o:P:S:")

func scpRun(cmd *cobra.Command, args []string) {
	var scpPath = config.MustLoad().GetString(config.CommandSCP)

	args, opts := sshParseArgs(scpFlags, args)

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
	a := []string{scpPath, "-3", fmt.Sprintf("-oProxyCommand=%s pipe %%h ssh", os.Args[0])}
	a = append(a, fmt.Sprintf("-oUserKnownHostsFile=%s", config.MustLoad().GetFilePath(config.PathKnownHosts)))

	a = append(a, opts...)
	a = append(a, args...)

	// ExecExternalCommand won't return (potential errors will cause logrus.Fatal() calls)
	internal.ExecExternalCommand(scpPath, a)
}

func scpValidate(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if strings.Contains(toComplete, ":") {
		return nil, cobra.ShellCompDirectiveDefault
	}

	return internal.DeviceListCompletion{}.Run(cmd, args, toComplete)
}
