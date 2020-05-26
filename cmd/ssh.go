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

// option list copied from debian jessie's openssh source package (from ssh.c, line 509)
var sshFlags = sshParseFlags("1246ab:c:e:fgi:kl:m:no:p:qstvxD:L:NR:")

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
	sshPath := "/usr/bin/ssh"

	// parse args (to detect the ones before 'user@host')
	args, opts := sshParseArgs(sshFlags, args)
	if len(args) < 1 {
		logrus.Fatal("missing target host")
	}

	// first non-option target is the [user@]host.
	tgtHost, tgtUser := sshParseTarget(args[0])
	args = args[1:]

	// compose ProxyCommand
	// TODO this will fail miserably if os.Args[0] contain spaces

	// compose the ssh command
	var a = make([]string, 0, 20)
	// ssh -oProxyCommand=ondevice pipe ssh %h ssh
	a = append(a, sshPath, fmt.Sprintf("-oProxyCommand=%s pipe %%h ssh", os.Args[0]))

	if sshGetConfig(opts, "UserKnownHostsFile") == "" {
		// use our own known_hosts file unless the user specified an override
		a = append(a, fmt.Sprintf("-oUserKnownHostsFile=%s", config.GetConfigPath("known_hosts")))
	}

	a = append(a, opts...) // ... ssh flags (-L -R -D ...)

	// target [user@]devId (qualified devId)
	if tgtUser != "" {
		a = append(a, fmt.Sprintf("%s@%s", tgtUser, tgtHost))
	} else {
		a = append(a, tgtHost)
	}
	a = append(a, args...) // non-option ssh arguments (command to be run on the host)

	// syscall.Exec will replace this app with ssh (yes, replace it, not just launch)
	// therefore, unless there's an error, this is the last line of code to be executed
	err := syscall.Exec(sshPath, a, os.Environ())
	if err != nil {
		logrus.WithError(err).Fatalf("failed to run '%s'", sshPath)
	}

	// nothing here should ever be executed
	logrus.Fatalf("this should never happen")
}

// sshGetConfig -- returns the specified -o SSH option (if present)
//
// note that key is case insensitive
func sshGetConfig(opts []string, key string) string {
	key = strings.ToLower(key)
	for _, opt := range opts {
		if !strings.HasPrefix(opt, "-o") {
			continue
		}
		var parts = strings.SplitN(opt[2:], "=", 2)
		if strings.ToLower(parts[0]) == key {
			return parts[1]
		}
	}

	return ""
}

// sshParseArgs -- Takes `ondevice ssh` arguments and parses them (into flags/options and other arguments)
func sshParseArgs(flags map[byte]bool, args []string) (outArgs []string, outOpts []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if len(arg) == 0 {
			logrus.Fatal("failed to parse ssh arguments: got an empty argument while looking for '[user@]host'")
		} else if arg == "-" {
			logrus.Fatal("stray '-' ssh argument")
		} else if arg[0] == '-' {
			hasValue, ok := flags[arg[1]]
			if !ok {
				logrus.Fatal("unsupported ssh argument: ", arg)
			}
			if hasValue && len(arg) == 2 {
				// the value's in the next argument, push them as one (simplifying subsequent parsing)
				outOpts = append(outOpts, arg+args[i+1])
				i++
			} else if hasValue && len(arg) > 2 {
				// the value's part of arg
				outOpts = append(outOpts, arg)
			} else if !hasValue && len(arg) > 2 { // && !hasValue
				logrus.Fatal("got value for flag that doesn't expect one: ", arg)
			} else if !hasValue && len(arg) == 2 {
				// a simple flag
				outOpts = append(outOpts, arg)
			} else {
				// yay to defensive programming
				logrus.Fatal("this should never happen (fifth state of two binary values)")
			}
		} else {
			// first non-option argument -> we're done
			outArgs = args[i:]
			break
		}
	}

	return
}

// sshParseFlags -- takes a getopt-style argument string and returns a map
// of flag characters and whether or not they expect an argument
func sshParseFlags(flags string) map[byte]bool {
	rc := map[byte]bool{}

	for i := 0; i < len(flags); i++ {
		flag := flags[i]
		hasValue := false
		if flags[i+1] == ':' {
			hasValue = true
			i++
		}
		rc[flag] = hasValue
	}
	return rc
}

func sshParseTarget(target string) (tgtHost string, tgtUser string) {
	parts := strings.SplitN(target, "@", 2)
	if len(parts) == 1 {
		tgtUser = ""
		tgtHost = parts[0]
	} else if len(parts) == 2 {
		tgtUser = parts[0]
		tgtHost = parts[1]
	} else {
		logrus.Fatal("SplitN(..., 2) returned odd number of results: ", len(parts))
	}

	// always use qualified device names
	if !strings.Contains(tgtHost, ".") {
		user, _, err := config.GetClientAuth()
		if err == nil {
			tgtHost = fmt.Sprintf("%s.%s", user, tgtHost)
		}
	}

	return tgtHost, tgtUser
}
