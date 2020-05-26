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
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func configRun(cmd *cobra.Command, args []string) {
	var keys = viper.AllKeys()
	sort.Strings(keys)
	for _, key := range keys {
		var val = viper.GetString(key)
		fmt.Printf("%s=%s\n", key, val)
	}
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "show or modify your local ondevice configuration",
	Long:  `calling ondevice config without parameters will print a list of key=value lines`,
	Run:   configRun,
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
	
  $ ondevice.config.get ssh.path rsync.path
  ssh.path=ssh
  rsync.path=rsync`,
	Run: func(cmd *cobra.Command, args []string) {
		var printKeyVal = len(args) > 1 // print in the form key=val if more than one key was specified
		var rc = 0

		for _, key := range args {
			var val = viper.Get(args[0])
			if val == nil {
				logrus.Error("config key not found: ", args[0])
				rc = 1
				continue
			}
			if printKeyVal {
				fmt.Printf("%s=%v\n", key, val)
			} else {
				fmt.Println(val)
			}
		}

		if rc != 0 {
			os.Exit(rc)
		}
	},
	Args: cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var matchingKeys []string
		var shellDirective = cobra.ShellCompDirectiveNoFileComp

		for _, k := range viper.AllKeys() {
			if strings.HasPrefix(k, toComplete) {
				matchingKeys = append(matchingKeys, k)
			}
		}

		return matchingKeys, shellDirective
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
}
