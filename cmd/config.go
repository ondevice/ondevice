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
	"sort"

	"github.com/ondevice/ondevice/cmd/internal"
	"github.com/ondevice/ondevice/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func configRun(cmd *cobra.Command, args []string) {
	var values = config.MustLoad().AllValues()

	var keys []string
	for k := range values {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s=%s\n", key, values[key])
	}
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "show or modify your local ondevice configuration",
	Long:  `calling ondevice config without parameters will print a list of key=value lines`,
	Run:   configRun,
	Args:  cobra.NoArgs,
	//Hidden: true,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "print individual configuration values",
	Long: `ondevice config get prints one or more configuration values.

when more than one key is specified, each line of output will contain 'key=value' pairs.
when only one key is requested, only the value will be printed`,
	Example: `  $ ondevice config get ssh.path
  ssh
	
  $ ondevice config get ssh.path rsync.path
  ssh.path=ssh
  rsync.path=rsync`,
	Run: func(cmd *cobra.Command, args []string) {
		var printKeyVal = len(args) > 1 // print in the form key=val if more than one key was specified
		var rc = 0

		var cfg = config.MustLoad()

		for _, keyName := range args {
			var key = config.FindKey(keyName)
			if key == nil {
				logrus.Errorf("config key not found: '%s'", keyName)
				rc = 1
				continue
			}

			var val = cfg.GetString(*key)

			if printKeyVal {
				fmt.Printf("%v=%s\n", key, val)
			} else {
				fmt.Println(val)
			}
		}

		if rc != 0 {
			os.Exit(rc)
		}
	},
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: internal.ConfigCompletion{WithReadOnly: true}.Run,
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>...",
	Short: "revert one or more configuration values to their defaults",
	Long:  `ondevice config unset deletes the specified configuration values.`,
	Example: `  $ ondevice config unset ssh.path rsync.path
  $ ondevice config unset client.timeout`,
	Run: func(cmd *cobra.Command, args []string) {
		var rc = 0

		var cfg = config.MustLoad()
		for _, arg := range args {
			var key = config.FindKey(arg)

			if key == nil {
				logrus.Fatalf("config key not found: '%s'", arg)
				rc = 1
				continue
			}

			cfg.Unset(*key)
		}

		if rc != 0 {
			os.Exit(rc)
		}

		if err := cfg.Write(); err != nil {
			logrus.WithError(err).Error("failed to write config")
		}
	},
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: internal.ConfigCompletion{WithReadOnly: false}.Run,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configUnsetCmd)
}
