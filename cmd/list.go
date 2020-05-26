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
	"encoding/json"
	"fmt"
	"os"

	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/filter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
type listCmd struct {
	cobra.Command

	jsonFlag     bool
	propsFlag    bool
	printIDsFlag bool
	stateFlag    string
	userFlag     string
}

func init() {
	var c listCmd
	c.Command = cobra.Command{
		Use:   "list [filters...]",
		Short: "list your devices",
		Long: `list your devices, optionally filtering the results

Filters:
  with filters you can limit the output based on device properties.

  filters have the format:
    propertyName[operator[value]]
  e.g.:
    somePackageVersion<2.3.4
    foo=

  All comparisons are string-based (e.g. "234" is less than "99")
  Missing values are treated like the empty string
  ("arch=" lists devices with empty or missing 'arch' property).

  The following comparison operators are supported (comma-separated):
    "=,==,!=,<,<<,<=,>,>>,>="

  The one-character operators ("=,<,>") each have a two-character alias ("==,<<,>>"
  respectively). These are provided to fix ambiguities (if the value to c start with "")
  Note that you might have to escape '>' and '<' to shell redirection`,
		Example: `  $ ondevice list
  ID            State   IP             Version         Name
  demo.7t91ta   offline                ondevice v0.4.3
  demo.fbqh2p   offline 192.168.1.23   ondevice v0.3.9
  demo.q5dkpm   online  127.0.0.1      ondevice v0.4.2
  demo.thm7br   offline 10.0.0.127     ondevice v0.4.3 My Raspberry PI

- using filters
  $ ondevice list 'fooVersion<2.3.4' 'foo=' --print-ids
  demo.7t91ta
  demo.fbqh2p
  demo.thm7br

  #only lists devices with the "fooVersion" property less than "2.3.4" (simple
  #string comparison, so "2.3.4" < "2.34.5") and without the "foo" property
  #(unset is equivalent to "")

- JSON output
  $ ondevice list --json --props
  {"id":"demo.7t91ta",state":"offline","stateTs":1490197318991,"version":"ondevice v0.4.3"}
  {"id":"demo.fbqh2p","ip":"192.168.1.23","state":"offline","stateTs":1485721709598,"version":"ondevice v0.3.9"}
  {"id":"demo.q5dkpm","ip":"127.0.0.1","state":"offline","stateTs":1487068641353,"version":"ondevice v0.4.2","props":{"test":"1234"}}
  {"id":"demo.thm7br","ip":"10.0.0.127","state":"offline", "": "My Raspberry PI","stateTs":1490963689912,"version":"ondevice v0.4.3"}

  #note that JSON fields may be missing or null`,
		Run: c.run,
	}
	rootCmd.AddCommand(&c.Command)

	c.Flags().BoolVar(&c.jsonFlag, "json", false, "output JSON, one line/object per device")
	c.Flags().BoolVar(&c.propsFlag, "props", false, "include properties (only affects JSON output)")
	c.Flags().BoolVar(&c.printIDsFlag, "print-ids", false, "print one devId per line (instead of tabular or JSON output)")

	c.Flags().StringVar(&c.stateFlag, "state", "", "limit to devices that are 'online'/'offline'")
	c.Flags().StringVar(&c.userFlag, "user", "", "show devices for a different user")
	c.Flags().MarkHidden("user")
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func (c *listCmd) run(cmd *cobra.Command, filters []string) {
	var err error

	// check output flags
	if c.jsonFlag && c.printIDsFlag {
		logrus.Fatal("specified conflicting output modes (--json and --print-ids)")
	}

	// --user
	var auth api.Authentication
	if c.userFlag != "" {
		if auth, err = api.GetClientAuthForUser(c.userFlag); err != nil {
			logrus.Fatalf("can't find client auth for user '%s'", c.userFlag)
			return
		}
	} else {
		if auth, err = api.GetClientAuth(); err != nil {
			logrus.Fatal("missing client auth, have you run 'ondevice login'?")
			return
		}
	}

	// imply --props if filters have been specified
	if len(filters) > 0 {
		c.propsFlag = true
	}

	allDevices, err := api.ListDevices(c.stateFlag, c.propsFlag, auth)
	if err != nil {
		logrus.WithError(err).Fatal("failed to fetch device list")
	}

	var devices = make([]api.Device, 0, len(allDevices))
	for _, dev := range allDevices {
		var ok bool
		if ok, err = c._matches(dev, filters); err != nil {
			logrus.WithError(err).Fatal("failed to filter device list")
		} else if ok {
			devices = append(devices, dev)
		}
	}

	if c.jsonFlag {
		c.printJSON(devices)
	} else if c.printIDsFlag {
		for _, dev := range devices {
			if _, err = fmt.Println(dev.ID); err != nil {
				logrus.WithError(err).Error("print failed")
				break
			}
		}
	} else {
		c.print(devices)
	}
}

func (c *listCmd) print(devices []api.Device) {
	// find the maximum lengths for each column
	titles := []string{"ID", "State", "IP", "Version", "Name"}
	widths := []int{2, 5, 2, 7, 4}
	for _, dev := range devices {
		cols := c._getColumns(dev)
		for j, col := range cols {
			width := len(col)
			if width > widths[j] {
				widths[j] = width
			}
		}
	}

	c._printColumns(widths, titles, os.Stderr)

	for _, dev := range devices {
		c._printColumns(widths, c._getColumns(dev), os.Stdout)
	}
}

func (*listCmd) printJSON(devs []api.Device) {
	for _, dev := range devs {
		out, err := json.Marshal(dev)
		if err != nil {
			logrus.WithError(err).Fatal("JSON serialization failed")
		}
		fmt.Println(string(out))
	}
}

func (*listCmd) _getColumns(dev api.Device) []string {
	return []string{dev.ID, dev.State, dev.IP, dev.Version, dev.Name}
}

// returns true if the device matches the given --with(out) flags
func (*listCmd) _matches(dev api.Device, filters []string) (bool, error) {
	for _, f := range filters {
		if ok, err := filter.Matches(dev, f); err != nil {
			return false, err
		} else if !ok {
			return false, nil
		}
	}

	return true, nil
}

func (c *listCmd) _printColumns(widths []int, cols []string, w *os.File) {
	if len(widths) != len(cols) {
		logrus.Fatalf("mismatch between cols and widths count (cols=%d, widths=%d)", cols, widths)
	}

	for i, width := range widths {
		c._printValue(width, cols[i], w)
		fmt.Fprint(w, " ")
	}
	fmt.Fprintln(w, "")
}

func (*listCmd) _printValue(width int, val string, w *os.File) {
	if len(val) > width {
		logrus.Fatal("width < len(val) !")
	}
	fmt.Fprint(w, val)
	for i := len(val); i < width; i++ {
		fmt.Fprint(w, " ")
	}
}
