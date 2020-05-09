package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// EventListener -- Listens for account events
type EventListener struct {
	// If set, only returns past events older than the specified eventId
	Until *int64
	// If set, also returns past events newer than (and equal to) the specified eventId
	Since *int64
	// Specifies the limit for past events (for Since and Until)
	Count *int
	// Timeout in seconds after which we exit
	Timeout *int
	// comma-separated list of event types
	Types string
	// comma-separated list of (unqualified) devIDs
	Devices string
}

// Event -- Represents an ondevice.io account event
type Event struct {
	ID     int64                  `json:"id"`
	Type   string                 `json:"type"`
	TS     int64                  `json:"ts"`
	User   string                 `json:"user"`
	Device string                 `json:"device,omitempty"`
	Msg    string                 `json:"msg"`
	Data   map[string]interface{} `json:"data"`
}

// Listen -- Listen for account events
func (e *EventListener) Listen(cb func(Event) error) error {
	var resp *http.Response
	var err error

	// prepare parameters
	var params = map[string]string{}
	if e.Until != nil {
		params["until"] = strconv.FormatInt(*e.Until, 10)
	}
	if e.Since != nil {
		params["since"] = strconv.FormatInt(*e.Since, 10)
	}
	if e.Count != nil && *e.Count >= 0 {
		params["count"] = strconv.FormatInt(int64(*e.Count), 10)
	}
	if e.Types != "" {
		params["types"] = e.Types
	}
	if e.Devices != "" {
		params["devices"] = e.Devices
	}
	if e.Timeout != nil && *e.Timeout >= 0 {
		params["timeout"] = strconv.FormatInt(int64(*e.Timeout), 10)
	}

	// send request
	if resp, err = get("/event.stream", params); err != nil {
		return err
	}

	reader := bufio.NewReader(resp.Body)
	var line []byte
	var eventType string
	var msg []byte

	for true {
		data, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				return err
			} else {
				break
			}
		}
		line = append(line, data...)
		if !isPrefix {
			// parse line
			//logrus.Debug("got event line: ", string(line))
			if len(line) == 0 {
				if len(eventType) > 0 && len(msg) > 0 {
					if eventType == "event" {
						var eventData Event
						if err = json.Unmarshal(msg, &eventData); err != nil {
							return err
						}
						//logrus.Debug("got event: ", eventData)
						if err = cb(eventData); err != nil {
							return err
						}
					} else if eventType == "ping" {
						logrus.Debug("got event stream ping: ", string(msg))
					} else {
						logrus.Warningf("unexpected event type: '%s' (msg: '%s')", eventType, msg)
					}
				}
				eventType = ""
				msg = []byte{}
			} else {
				parts := bytes.SplitN(line, []byte(": "), 2)
				if len(parts) < 2 {
					return fmt.Errorf("Missing ':' in EventSource: '%s'", line)
				}
				if bytes.Equal(parts[0], []byte("event")) {
					eventType = string(parts[1])
				} else if bytes.Equal(parts[0], []byte("data")) {
					if len(msg) > 0 {
						msg = append(msg, '\n')
					}
					msg = append(msg, parts[1]...)
				}
			}
			line = []byte{}
		}
	}

	return nil
}
