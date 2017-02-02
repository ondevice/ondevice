package tunnel

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ondevice/ondevice/logg"
	"github.com/ondevice/ondevice/util"
)

// Tunnel -- an ondevice.io tunnel ()
type Tunnel struct {
	Connection
	Side      string // "client" or "device"
	connected chan error
	wdog      *util.Watchdog // the client will use this to periodically send 'meta:ping' messages, the device will respond (and kick the Watchdog in the process)
	lastPing  time.Time

	readEOF, writeEOF bool

	CloseListeners   []func()
	DataListeners    []func(data []byte)
	EOFListeners     []func()
	TimeoutListeners []func()
}

// GetErrorCodeName -- returns a string representing the given 'HTTP-ish' tunnel error code
func GetErrorCodeName(code int) string {
	switch code {
	case 400:
		return "Bad Request"
	case 403:
		return "Access Denied"
	case 404:
		return "Not Found"
	case 503:
		return "Service Unavailable"
	}

	return ""
}

// CloseWrite -- send an EOF to the remote end of the tunnel (i.e. close the write channel)
func (t *Tunnel) CloseWrite() {
	logg.Debug("Sending EOF...")
	t.SendBinary([]byte("meta:EOF"))
}

func (t *Tunnel) Write(data []byte) {
	var msg = make([]byte, 0, len(data)+5)
	msg = append(msg, []byte("data:")...)
	msg = append(msg, data...)

	t.SendBinary(msg)
}

func (t *Tunnel) onMessage(_type int, msg []byte) {
	parts := bytes.SplitN(msg, []byte(":"), 2)

	if _type != websocket.BinaryMessage {
		logg.Error("Got non-binary message over the tunnel")
		return
	}
	if len(parts) < 2 {
		logg.Error("Missing colon in tunnel message")
		return
	}
	msgType := string(parts[0])
	msg = parts[1]

	if msgType == "meta" {
		parts = bytes.SplitN(msg, []byte(":"), 2)
		metaType := string(parts[0])

		if metaType == "ping" {
			//logg.Debug("got tunnel ping")
			pong := []byte("meta:pong")
			t.lastPing = time.Now()
			t.wdog.Kick()

			if len(parts) > 1 {
				pong = append(pong, byte(':'))
				pong = append(pong, msg[5:]...)
			}
			t.SendBinary(pong)
		} else if metaType == "pong" {
			logg.Debug("got tunnel pong: ", string(msg))
			t.lastPing = time.Now()
		} else if metaType == "connected" {
			logg.Debug("connected")
			t.connected <- nil
		} else if metaType == "EOF" {
			t.readEOF = true

			// call listeners
			for _, cb := range t.EOFListeners {
				cb()
			}
		} else {
			t._error(fmt.Errorf("Unsupported meta message: %s", metaType))
		}
	} else if msgType == "data" {
		if len(t.DataListeners) == 0 {
			panic("Tunnel: Missing OnData handler")
		}

		// call listeners
		for _, cb := range t.DataListeners {
			cb(msg)
		}
	} else if msgType == "error" {
		parts := strings.SplitN(string(msg), ":", 2)
		var code int
		var errMsg string
		if len(parts) == 1 {
			errMsg = parts[0]
		} else {
			code, _ = strconv.Atoi(parts[0])
			errMsg = parts[1]
		}

		err := fmt.Errorf("%s (%d): %s", GetErrorCodeName(code), code, errMsg)
		if t.connected != nil {
			t.connected <- err
		}
		t._error(err)
	} else {
		logg.Warning("Unsupported tunnel message type: ", msgType)
	}
}

func (t *Tunnel) _error(err error) {
	if len(t.ErrorListeners) == 0 {
		logg.Error(err)
	}
	for _, cb := range t.ErrorListeners {
		cb(err)
	}
}

func (t *Tunnel) _onTimeout() {
	for _, cb := range t.TimeoutListeners {
		cb()
	}
}
