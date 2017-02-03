package tunnel

import (
	"bytes"
	"log"
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
	connected chan util.APIError
	wdog      *util.Watchdog // the client will use this to periodically send 'meta:ping' messages, the device will respond (and kick the Watchdog in the process)
	lastPing  time.Time

	readEOF, writeEOF       bool
	bytesRead, bytesWritten int64

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

// SendEOF -- send an EOF to the remote end of the tunnel (i.e. close the write channel)
func (t *Tunnel) SendEOF() {
	logg.Info("sending EOF")
	if t.writeEOF == true {
		logg.Warning("Attempting to close already closed write channel")
		return
	}

	logg.Debug("Sending EOF...")
	t.writeEOF = true
	t.SendBinary([]byte("meta:EOF"))
	t._checkClose()
}

func (t *Tunnel) Write(data []byte) {
	msg := append([]byte("data:"), data...)

	t.bytesWritten += int64(len(data))
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
			logg.Info("Got EOF")
			t.readEOF = true

			t._checkClose()

			// call listeners
			for _, cb := range t.EOFListeners {
				cb()
			}
		} else {
			t._error(util.NewAPIError(util.OtherError, "Unsupported meta message: ", metaType))
		}
	} else if msgType == "data" {
		t.bytesRead += int64(len(msg))

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

		err := util.NewAPIError(code, errMsg)
		if t.connected != nil {
			t.connected <- err
		}
		t._error(err)
	} else {
		logg.Warning("Unsupported tunnel message type: ", msgType)
	}
}

// _checkClose -- after sending/receiving EOF this method checks if the tunnel
// should be closed
func (t *Tunnel) _checkClose() {
	if t.readEOF && t.writeEOF {
		logg.Debug("EOF on both channels, closing tunnel - side: ", t.Side)
		if t.Side == "device" {
			time.AfterFunc(10*time.Second, t.Close)
		} else if t.Side == "client" {
			t.Close()
		} else {
			logg.Warning("Unsupported tunnel side: ", t.Side)
		}
	}
}

func (t *Tunnel) _error(err util.APIError) {
	if len(t.ErrorListeners) == 0 {
		logg.Error(err)
	}
	for _, cb := range t.ErrorListeners {
		cb(err)
	}
}

func (t *Tunnel) _onClose() {
	// print log message and stop timers
	// TODO stop timers
	if t.Side == "client" {
		logg.Debugf("Tunnel closed, bytesRead=%d, bytesWritten=%d", t.bytesRead, t.bytesWritten)
	} else if t.Side == "device" {
		logg.Infof("Connection closed, bytesRead=%d, bytesWritten=%d", t.bytesRead, t.bytesWritten)
	}
}

func (t *Tunnel) _onTimeout() {
	for _, cb := range t.TimeoutListeners {
		cb()
	}
}
