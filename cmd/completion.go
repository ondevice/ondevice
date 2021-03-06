/*
Copyright © 2020 ondevice.io <dev@ondevice.io>

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
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "generates bash completion scripts",
	Long: `to load completion run

. <(ondevice completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(ondevice completion)
`,
	Run: func(cmd *cobra.Command, args []string) {
		// if we've been started as 'on', update rootCmd.Use to reflect that (just a little cheat for myself tbh)
		if len(os.Args) > 0 && os.Args[0] == "on" {
			rootCmd.Use = "on"
		}

		rootCmd.GenBashCompletion(os.Stdout)
	},
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
