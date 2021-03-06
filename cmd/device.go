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
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/cmd/internal"
	"github.com/ondevice/ondevice/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deviceCmd represents the device command
type deviceCmd struct {
	cobra.Command

	yesFlag bool
}

func init() {
	var c deviceCmd
	c.Command = cobra.Command{
		Use:   "device <devId> <props/set/rm> [key1=val1 ...]",
		Short: "list/manipulate device properties",
		Long: `manage your devices' properties.

Properties can be used to keep track of your devices, to manage their characteristics,
keep tracks of running maintenance scripts, etc.

- list properties
  $ ondevice device <devId> props
- add/update properties
  $ ondevice device <devId> set [key1=val1 ...]
- remove properties
  $ ondevice device <devId> rm [--yes/-y] [key1 key2 ...]

Each invocation will print the resulting property list.
This command requires a client key with the 'manage' permission.

Have a look at ondevice list's filters for a simple way to list devices with specific properties.


Special properties:
- rename (= change devId of) a device
  $ ondevice device $devId set on:id=$newId

- remove the special property 'on:id' will attempt to delete the device
  (will ask for confirmation unless you also specify --yes)
  Only devices that have been offline for at least an hour can be deleted.
  $ ondevice device $devId rm on:id`,
		Example: `  $ ondevice device q5dkpm props
  $ ondevice device q5dkpm set test=1234 foo=bar
  test=1234
  foo=bar
  $ ondevice device q5dkpm rm foo
  test=1234

  # rename and then delete the device (using the on:id special property)
  $ ondevice device q5dkpm set on:id=rpi
  $ ondevice device rpi rm on:id
  Do you really want to delete the device 'rpi' (y/N):`,
		Run:               c.run,
		ValidArgsFunction: c.validateArgs,
	}

	rootCmd.AddCommand(&c.Command)

	c.Flags().BoolVarP(&c.yesFlag, "yes", "y", false, `don't ask before deleting a device (by removing the 'on:id' property).
Has no effect on other properties.`)
}

func (c *deviceCmd) run(_ *cobra.Command, args []string) {
	var err error

	if len(args) < 1 {
		err = errors.New("missing deviceId")
	} else if len(args) < 2 {
		err = errors.New("missing device command")
	}

	if err != nil {
		logrus.WithError(err).Fatal("error")
		return
	}

	var devID = args[0]
	var cmd = args[1]
	args = args[2:]

	var auth config.Auth
	if auth, err = config.LoadAuth().GetClientAuthForDevice(devID); err != nil {
		logrus.WithError(err).Fatal("missing client auth!")
	}

	switch cmd {
	case "set":
		err = c.setProperties(devID, args, auth)
	case "rm":
		err = c.removeProperties(devID, args, auth)
	case "props", "properties", "list":
		err = c.listProperties(devID, args, auth)
	default:
		err = fmt.Errorf("Unknown device command: '%s'", cmd)
	}

	if err != nil {
		logrus.WithError(err).Fatal("error")
		return
	}
	return
}

func (c *deviceCmd) listProperties(devID string, extraArgs []string, auth config.Auth) error {
	if len(extraArgs) > 0 {
		return errors.New("Too many arguments")
	}
	return c.printProperties(api.ListProperties(devID, auth))
}

func (c *deviceCmd) removeProperties(devID string, args []string, auth config.Auth) error {
	if len(args) == 0 {
		logrus.Error("too few arguments")
	}

	// check if the user wants to delete the device ("on:id" present)
	var wantsDelete = false
	for _, key := range args {
		if key == "on:id" {
			wantsDelete = true
			break
		}
	}

	if wantsDelete {
		if len(args) != 1 {
			logrus.Fatal("to delete a device, remove its 'on:id' property (and nothing else)")
		}

		var confirmed = c.yesFlag
		if !confirmed {
			var reader = bufio.NewReader(os.Stdin)
			var input string
			var err error

			for input == "" {
				fmt.Printf("Do you really want to delete the device '%s' (y/N): ", devID)
				input, err = reader.ReadString('\n')
				if err != nil {
					logrus.WithError(err).Fatal("failed to read your response")
				}

				switch strings.TrimSpace(strings.ToLower(input)) {
				case "y", "yes":
					confirmed = true
				case "n", "no", "":
					confirmed = false
					input = "no"
				default:
					input = ""
				}
			}
		}

		if confirmed {
			if err := api.DeleteDevice(devID, auth); err != nil {
				return err
			}
		} else {
			return errors.New("Aborted delete")
		}
		return nil
	}

	return c.printProperties(api.RemoveProperties(devID, args, auth))
}

func (c *deviceCmd) setProperties(devID string, args []string, auth config.Auth) error {
	var props = make(map[string]string)

	for _, arg := range args {
		s := strings.SplitN(arg, "=", 2)
		if _, ok := props[s[0]]; ok {
			return fmt.Errorf("Duplicate value for property '%s'", s[0])
		}
		props[s[0]] = s[1]
	}

	return c.printProperties(api.SetProperties(devID, props, auth))
}

func (c *deviceCmd) printProperties(props map[string]interface{}, err error) error {
	if err != nil {
		return err
	}

	// get list of keys and sort them (by scope)
	var sortedKeys = map[string][]string{}
	for k := range props {
		var scope = ""
		if parts := strings.SplitN(k, ":", 2); len(parts) == 2 {
			scope = parts[0]
			//k = parts[1] -- we'll keep the qualified key names
		}
		sortedKeys[scope] = append(sortedKeys[scope], k)
	}

	// sort scopes and the keys of each
	var sortedScopes = make([]string, 0, len(sortedKeys))
	for s := range sortedKeys {
		sort.Strings(sortedKeys[s])
		if s != "" {
			sortedScopes = append(sortedScopes, s)
		}
	}
	sort.Strings(sortedScopes)
	sortedScopes = append(sortedScopes, "") // unscoped properties go last

	for _, s := range sortedScopes {
		for _, k := range sortedKeys[s] {
			var v = props[k]
			var repr string

			if s, ok := v.(string); ok {
				repr = s
			} else {
				var reprBytes, _ = json.Marshal(v)
				repr = string(reprBytes)
			}

			fmt.Printf("%s=%s\n", k, repr)
		}
	}

	return nil
}

// validateArgs -- does shell completion
func (c deviceCmd) validateArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		// first arg -> devId
		return internal.DeviceListCompletion{
			DontIgnoreUser: true,
		}.Run(cmd, args, toComplete)
	} else if len(args) == 1 {
		return []string{"props", "set", "rm"}, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveDefault
}
