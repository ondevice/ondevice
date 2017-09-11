package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"git.coding.zone/ondevice/goserver/common/util"

	flags "github.com/jessevdk/go-flags"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/logg"
)

const _longEventHelp = `
ondevice event [options]
  Subscribe to your account's event stream.
ondevice event --until=<eventId> [--count=50]
  List past events up until the given eventId (useful for paging)

Options:
--json
  Prints the events as JSON objects, one per line
--since=eventId
  Specify the last eventId you've seen when invoking 'ondevice event' the last
  time.
  The event with the given ID will be included in the output (unless there have
  been more than --count events since then)
--until=eventId
  Only list past events, up until the given eventId (exits immediately)
  Can't be used in conjunction with --since, --timeout or --await
--count=n
  Display n existing events for --since or --until.
  Defaults to 50
--type=eventType[,...]
  Filters the output by event type (comma-separated list).
  Some types: deviceOnline, deviceOffline, connect, accept, close,
  For a full list of event types, have a look at the ondevice.io documentation.

--device=devId[,...]
  Filters the output for one or more devices (comma-separated list)
--timeout=n
  Stops the event stream after n seconds.
  0 means 'exit immediately' (will only print existing events), negative values
  disable timeouts.
  Exits with code 2.
  (To start where you left off, use the --since option)
--await=eventType
  Waits for an event of type eventType to happen (and exits with code 0 as soon
  as such an event was received).
  If both --timeout and --await are present, whichever one happens first will
  cause the program to exit (check the return code to see what happened first).
  If --since was specified, that event will be printed but won't trigger an exit


Examples:
  ondevice event --json --timeout=30 --since=1234
    List events for 30 seconds (you could run this in a loop, )
  ondevice event --json --device=dev1,dev2 --await=deviceOnline
    Exit as soon as one of the specified devices comes online (have a look at
    the output to see which one it is)
  ondevice event --count=50 --timeout=0
    List the 50 most recent events (and exit immediately)
  ondevice event --until=1234 --count=50
    List event 1234 and the 50 events before it (and exit immediately)

`

var errAwaitMatch = errors.New("Await match")

// EventsCommand -- Prints (or waits for) device events to happen
type EventsCommand struct {
	JSON    bool   `long:"json" description:"Print output in JSON format, one event per line"`
	Since   *int64 `long:"since" description:"List past events newer than the given eventId"`
	Until   *int64 `long:"until" description:"List past events older than the given eventId"`
	Count   *int   `long:"count" description:"Limit the number of past events"`
	Type    string `long:"type" description:"filter for events of the given type(s) (comma-separated)"`
	Device  string `long:"device" description:"filter for events of the given device(s) (comma-separated)"`
	Timeout *int   `long:"timeout" description:"Exit with code 2 after n seconds (0: exit immediately, default: no timeout)"`
	Await   string `long:"await" description:"Exit after receiving an event of the specified type"`

	timeoutWdog *util.Watchdog
}

func (e *EventsCommand) args() string {
	return "[--json] [--since=eventId] [--until=eventId] [--count=50] [--type=...] [--device=...] [--timeout=30] [--await=eventType]"
}

func (e *EventsCommand) longHelp() string {
	return _longEventHelp
}

func (e *EventsCommand) shortHelp() string {
	return "Prints past (and listens for live) account events"
}

func (e *EventsCommand) run(args []string) int {
	var err error

	if args, err = flags.ParseArgs(e, args); err != nil {
		logg.Fatal(err)
	}
	if len(args) > 0 {
		logg.Fatal("Too many arguments")
	}

	// argument checks
	if e.Until != nil && e.Since != nil {

	}

	// init listener
	listener := api.EventListener{
		Until:   e.Until,
		Since:   e.Since,
		Count:   e.Count,
		Devices: e.Device,
		Types:   e.Type,
		Timeout: e.Timeout,
	}

	if e.Timeout != nil {
		e.timeoutWdog = util.NewWatchdog(time.Duration(*e.Timeout)*time.Second, e.onTimeout)
	}

	if err := listener.Listen(e.onEvent); err != nil {
		if err == errAwaitMatch {
			// return 0
		} else {
			logg.Fatal(err)
		}
	}
	return 0
}

func (e *EventsCommand) onEvent(ev api.Event) error {
	// print event
	if e.JSON {
		data, err := json.Marshal(ev)
		if err != nil {
			logg.Fatal(err)
		}
		fmt.Println(string(data))
	} else {
		fmt.Printf("%s (id: %d): \t%s\n", util.MsecToTs(ev.TS).Format("2006/01/02 15:04:05"), ev.ID, ev.Msg)
	}

	// check 'await'
	if e.Await != "" && ev.Type == e.Await {
		if e.Since == nil || *e.Since < ev.ID {
			// found a match -> exit with code 0
			return errAwaitMatch
		}
	}

	return nil
}

func (e *EventsCommand) onTimeout() {
	// TODO exit gracefully (closing the listener etc.)
	logg.Info("event stream timeout")
	os.Exit(2)
}
