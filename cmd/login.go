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
	"fmt"
	"os"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "log in to the ondevice.io service",
	Long:  `log in to the ondevice.io service using one of your API keys.`,
	Example: `  $ ondevice login
  User: <enter your user name>
  Auth: <enter your credentials>`,
	Run: loginRun,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("batch", "", `Run in batch mode, using the given username and reading the authentication key
from stdin, e.g.:
  echo '5h42l5xylznw'|ondevice login --batch=demo`)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func loginRun(cmd *cobra.Command, args []string) {
	var user, auth string
	var err error

	user = cmd.Flag("batch").Value.String()

	reader := bufio.NewReader(os.Stdin)

	if user != "" {
		// run in batch mode
		if auth, err = reader.ReadString('\n'); err != nil {
			logrus.WithError(err).Fatal("failed to read auth key from stdin")
		}
		auth = strings.TrimSpace(auth)
	} else {
		fmt.Print("Please login using one of your ondevice.io Auth Keys.\n\n")

		fmt.Println("-----")
		fmt.Println("DO NOT use your account password - it won't work.")
		fmt.Println("If you haven't set one up yet, visit https://my.ondevice.io/me/keys")
		fmt.Println("(see https://docs.ondevice.io/basics/auth-keys/ for details)")
		fmt.Print("-----\n\n")

		fmt.Print("User: ")
		user, err = reader.ReadString('\n')
		if err != nil {
			logrus.WithError(err).Fatal("failed to read user name")
		}
		user = strings.TrimSpace(user)

		fmt.Printf("Auth key: ")
		var authBytes []byte
		if authBytes, err = gopass.GetPasswd(); err != nil {
			logrus.WithError(err).Fatal("failed to read auth key")
		}
		auth = string(authBytes)
	}

	info, err := api.GetKeyInfo(config.NewAuth(user, auth))
	if err != nil {
		logrus.WithError(err).Fatal("failed to verify login info")
	}

	if info.Key != "" {
		// the API server wants us to use a different auth key
		// (most likely because the user has used their account password)
		auth = info.Key
	}

	// display any messages the server might have for us
	for _, msg := range info.Messages {
		var parts = strings.SplitN(msg, ":", 2)
		switch parts[0] {
		case "info":
			logrus.Info(parts[1])
		case "warn":
			logrus.Warning(parts[1])
		case "err":
			logrus.Error(parts[1])
		default:
			logrus.Info("Got server message: ", msg)
		}
	}

	// update auth
	if info.IsType("client") {
		logrus.Info("updating client auth")
		if err := config.SetAuth("client", user, string(auth)); err != nil {
			logrus.WithError(err).Fatal("failed to set client auth")
		}
	}
	if info.IsType("device") {
		logrus.Info("updating device auth")
		if err := config.SetAuth("device", user, string(auth)); err != nil {
			logrus.WithError(err).Fatal("failed to set device auth")
		}
	}
}
