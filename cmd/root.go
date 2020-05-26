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
	"path"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ondevice",
	Short: "ssh into your devices even if they don't have a public IP",
	Long: `ondevice wraps your favourite ssh-based tools (ssh, sftp, rsync, etc.)
to give you access to all your devices, even if they're in another network.`,
	Example: `- initially log in to your account
  $ ondevice login

- connect to your NAS (i.e. device named 'nas')
  $ ondevice ssh nas

- run apt-get on your office pc
  $ ondevice ssh office apt-get -y install vim

- copy files using rsync (requires rsync to be installed on both hosts)
  $ ondevice rsync -av office:~/Documents/Project-XY ~/Documents/Work-Stuff/

- set up port forwarding to a psql server that only accepts local connections
  $ ondevice ssh -L 54320:localhost:5432 -f -N user@webserver
  $ psql -h localhost -p 54320 -U webapp -W webapp`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ondevice.yaml)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "conf", "", "alias for '--config'")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".ondevice" (without extension).
		viper.AddConfigPath(path.Join(home, ".config/ondevice"))
		viper.SetConfigFile(path.Join(home, ".config/ondevice/ondevice.conf"))
		viper.SetConfigType("ini")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Info("using config file: ", viper.ConfigFileUsed())
	} else {
		logrus.WithError(err).Error("error reading config")
	}

	/*	for k, v := range viper.AllSettings() {
		fmt.Printf("- config: %s=%v\n", k, v)
	}*/
}
