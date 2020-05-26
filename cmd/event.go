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
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var errAwaitMatch = errors.New("Await match")

type eventCmd struct {
	cobra.Command

	jsonFlag    bool
	sinceFlag   int64
	untilFlag   int64
	countFlag   int
	typeFlag    string
	deviceFlag  string
	timeoutFlag int
	awaitFlag   string

	visitedFlags map[string]int

	timeoutWdog *util.Watchdog
}

// eventCmd represents the event command

func init() {
	var c eventCmd
	c.visitedFlags = make(map[string]int)
	c.Command = cobra.Command{
		Use:   "event",
		Short: "prints past (and listens for live) account events",
		Long:  `subscribe to your account's event stream and/or list past events.`,
		Example: `- list past events up until the given eventId (useful for paging)
  $ ondevice event --until=<eventId> [--count=50]

- list events for 30 seconds (you could run this in a loop and do other stuff every 30sec)
  $ ondevice event --json --timeout=30 --since=1234

- exit as soon as one of the specified devices comes online
  (have a look at the output to see which one it is)
  $ ondevice event --json --device=dev1,dev2 --await=deviceOnline

- list the 50 most recent events (and exit immediately)
  $ ondevice event --count=50 --timeout=0
  
- list event 1234 and the 50 events before it (and exit immediately)
  $ ondevice event --until=1234 --count=50`,
		Run:  c.run,
		Args: cobra.NoArgs,
	}
	rootCmd.AddCommand(&c.Command)

	c.Flags().BoolVar(&c.jsonFlag, "json", false, "print output in JSON format, one event per line")
	c.Flags().Int64Var(&c.sinceFlag, "since", -1, `if specified, only list events newer than the given eventId.
specify the last eventId you've seen when invoking 'ondevice event' the last time.
The event with the given ID will be included in the output unless there have been more than --count events since then`)
	c.Flags().Int64Var(&c.untilFlag, "until", -1, `If specified, only list events older than the given eventId.
Only list past events, up until the given eventId (exits immediately)
Can't be used in conjunction with --since, --timeout or --await`)
	c.Flags().IntVar(&c.countFlag, "count", 50, `limit the number of past events returned for --since or --until.
Doesn't affect live events printed after that.`)
	c.Flags().StringVar(&c.typeFlag, "type", "", `Filters the output by event type (comma-separated list).
Some types: deviceOnline, deviceOffline, connect, accept, close,
For a full list of event types, have a look at the ondevice.io documentation.`)
	c.Flags().StringVar(&c.deviceFlag, "device", "", `only show events for the given device(s) (comma-separated)`)
	c.Flags().IntVar(&c.timeoutFlag, "timeout", -1, `exit with code 2 after n seconds (0: exit immediately, default: no timeout)
Stops the event stream after n seconds.
0 means 'exit immediately' (will only print existing events), negative values
disable timeouts.
Exits with code 2, can be used in conjunction with --await.
To start where you left off, use the --since option`)
	c.Flags().StringVar(&c.awaitFlag, "await", "", `exit with code 0 after receiving an event of the specified type
If both --timeout and --await are present, whichever one happens first will
cause the program to exit (check the return code to see what happened first).`)
}

func (c *eventCmd) run(cmd *cobra.Command, args []string) {
	// init listener
	listener := api.EventListener{
		Devices: c.deviceFlag,
		Types:   c.typeFlag,
	}
	if c.flagWasSet("since") {
		listener.Since = &c.sinceFlag
	}
	if c.flagWasSet("until") {
		listener.Until = &c.untilFlag
	}
	if c.flagWasSet("count") {
		listener.Count = &c.countFlag
	}

	if c.flagWasSet("timeout") {
		listener.Timeout = &c.timeoutFlag
		c.timeoutWdog = util.NewWatchdog(time.Duration(c.timeoutFlag)*time.Second, c.onTimeout)
	}

	// default timeout (set in ondevice.go) is 30sec.
	// this can be long-running by design -> reset timeout
	http.DefaultClient.Timeout = 0
	if err := listener.Listen(c.onEvent); err != nil {
		if err == errAwaitMatch {
			// return 0
		} else {
			logrus.WithError(err).Fatal("error")
		}
	}
}

func (c *eventCmd) onEvent(ev api.Event) error {
	// print event
	if c.jsonFlag {
		data, err := json.Marshal(ev)
		if err != nil {
			logrus.WithError(err).Fatal("failed to marshal event")
		}
		fmt.Println(string(data))
	} else {
		fmt.Printf("%s (id: %d): \t%s\n", util.MsecToTs(ev.TS).Format("2006/01/02 15:04:05"), ev.ID, ev.Msg)
	}

	// check 'await'
	if c.awaitFlag != "" && ev.Type == c.awaitFlag {
		if !c.flagWasSet("since") || c.sinceFlag < ev.ID {
			// found a match -> exit with code 0
			return errAwaitMatch
		}
	}

	return nil
}

func (*eventCmd) onTimeout() {
	// TODO exit gracefully (closing the listener etc.)
	logrus.Info("event stream timeout")
	os.Exit(2)
}

func (c *eventCmd) flagWasSet(name string) bool {
	if f := c.Flag(name); f != nil {
		return f.Changed
	}
	logrus.Error("eventCmd: unexpected flag: ", name)
	return false
}
