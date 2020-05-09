package tunnel

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ondevice/ondevice/util"
	"github.com/sirupsen/logrus"
)

// Tunnel -- an ondevice.io tunnel ()
type Tunnel struct {
	Connection

	// Side -- either DeviceSide or ClientSide
	Side      string
	connected chan util.APIError
	wdog      *util.Watchdog // the client will use this to periodically send 'meta:ping' messages, the device will respond (and kick the Watchdog in the process)
	lastPing  time.Time

	readEOF, writeEOF bool

	// metrics:
	bytesRead, bytesWritten int64
	startTs                 time.Time

	// listeners
	DataListeners    []func(data []byte)
	EOFListeners     []func()
	TimeoutListeners []func()
}

const (
	// ClientSide -- This Tunnel instance represents the client side of the tunnel (see Tunnel.Side)
	ClientSide = "client"
	// DeviceSide -- This Tunnel instance represents the device side of the tunnel (see Tunnel.Side)
	DeviceSide = "device"
)

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

func (t *Tunnel) _initTunnel(side string) {
	t.connected = make(chan util.APIError)
	t.Side = side
	t.startTs = time.Now()
	t.CloseListeners = append([]func(){t._onClose}, t.CloseListeners...)
}

// SendEOF -- send an EOF to the remote end of the tunnel (i.e. close the write channel)
func (t *Tunnel) SendEOF() {
	if t.writeEOF == true {
		logrus.Debug("Attempted to close already closed write channel")
		return
	}

	logrus.Debug("sending EOF")
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
		logrus.Error("got non-binary message over the tunnel")
		return
	}
	if len(parts) < 2 {
		logrus.Error("missing colon in tunnel message")
		return
	}
	msgType := string(parts[0])
	msg = parts[1]

	if msgType == "meta" {
		parts = bytes.SplitN(msg, []byte(":"), 2)
		metaType := string(parts[0])

		if metaType == "ping" {
			//logrus.Debug("got tunnel ping")
			pong := []byte("meta:pong")
			t.lastPing = time.Now()
			t.wdog.Kick()

			if len(parts) > 1 {
				pong = append(pong, byte(':'))
				pong = append(pong, msg[5:]...)
			}
			t.SendBinary(pong)
		} else if metaType == "pong" {
			logrus.Debug("got tunnel pong: ", string(msg))
			t.lastPing = time.Now()
		} else if metaType == "connected" {
			logrus.Debug("connected")
			if err := t.state.Event("connected"); err != nil {
				logrus.WithError(err).Error("state change failed (ev: 'connected')")
			}
			t.connected <- nil
		} else if metaType == "EOF" {
			t._onEOF()
			t._checkClose()
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
		logrus.Warning("unsupported tunnel message type: ", msgType)
	}
}

// _checkClose -- after sending/receiving EOF this method checks if the tunnel
// should be closed
func (t *Tunnel) _checkClose() {
	if t.readEOF && t.writeEOF {
		logrus.Debug("EOF on both channels, closing tunnel - side: ", t.Side)
		if t.Side == DeviceSide {
			// it's the client's job to actually close the tunnel - but if it doesn't
			// do that in time, we'll do it ourselves
			time.AfterFunc(10*time.Second, t.Close)
		} else if t.Side == ClientSide {
			t.Close()
		} else {
			logrus.Warning("unsupported tunnel side: ", t.Side)
		}
	}
}

func (t *Tunnel) _error(err util.APIError) {
	if len(t.ErrorListeners) == 0 {
		logrus.WithError(err).Error("tunnel error")
	}
	for _, cb := range t.ErrorListeners {
		cb(err)
	}
}

func (t *Tunnel) _onClose() {
	t.writeEOF = true // no need to send an EOF over a closed tunnel
	t._onEOF()        // always fire the EOF signal

	// print log message and stop timers
	duration := time.Now().Sub(t.startTs)
	msg := fmt.Sprintf("Tunnel closed, bytesRead=%d, bytesWritten=%d, duration=%s", t.bytesRead, t.bytesWritten, duration.String())
	if t.Side == ClientSide {
		logrus.Debug(msg)
	} else if t.Side == DeviceSide {
		logrus.Info(msg)
	}

	// TODO stop timers
}

func (t *Tunnel) _onEOF() {
	logrus.Debug("Tunnel._onEOF()")
	if t.readEOF == true {
		return
	}
	t.readEOF = true

	// call listeners
	for _, cb := range t.EOFListeners {
		cb()
	}
}

func (t *Tunnel) _onTimeout() {
	for _, cb := range t.TimeoutListeners {
		cb()
	}
}
