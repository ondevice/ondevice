package tunnel

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

// Tunnel -- an ondevice.io tunnel ()
type Tunnel struct {
	Connection
	Side      string // "client" or "device"
	connected chan error

	OnClose func()
	OnData  func(data []byte)
	OnError func(err error)
	OnEOF   func()
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
	log.Print("Sending EOF...")
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
		log.Fatal("Got non-binary message over the tunnel")
	}
	if len(parts) < 2 {
		log.Fatal("Missing colon in tunnel message")
	}
	msgType := string(parts[0])
	msg = parts[1]

	if msgType == "meta" {
		parts = bytes.SplitN(msg, []byte(":"), 2)
		metaType := string(parts[0])

		if metaType == "ping" {
			pong := []byte("meta:pong")
			if len(parts) > 1 {
				pong = append(pong, byte(':'))
				pong = append(pong, parts[1]...)
			}
			t.SendBinary(pong)
			// TODO update 't.lastPing'
		} else if metaType == "pong" {
			// TODO update 't.lastPing'
		} else if metaType == "connected" {
			log.Print("connected")
			t.connected <- nil
		} else if metaType == "EOF" {
			if t.OnEOF != nil {
				t.OnEOF()
			}
		} else {
			t._error(fmt.Errorf("Unsupported meta message: %s", metaType))
		}
	} else if msgType == "data" {
		if t.OnData == nil {
			panic("Tunnel: Missing OnData handler")
		}
		t.OnData(msg)
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
		log.Println("Unsupported tunnel message type: ", msgType)
	}
}

func (t *Tunnel) _error(err error) {
	if t.OnError != nil {
		t.OnError(err)
	} else {
		log.Print("Error: ", err)
	}
}
