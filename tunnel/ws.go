package tunnel

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
	"github.com/ondevice/ondevice/util"
)

// Connection -- WebSocket connection
type Connection struct {
	ws       *websocket.Conn
	isClosed bool

	CloseListeners    []func()
	ErrorListeners    []func(err util.APIError)
	MessageListerners []func(int, []byte)

	writeLock sync.Mutex
	done      chan struct{}
}

// OpenWebsocket -- Open a websocket connection
func OpenWebsocket(c *Connection, endpoint string, params map[string]string, onMessage func(int, []byte), auths ...api.Authentication) util.APIError {
	hdr := http.Header{}

	var auth api.Authentication
	if len(auths) == 0 {
		var err error
		if auth, err = api.CreateClientAuth(); err != nil {
			return util.NewAPIError(util.OtherError, err.Error())
		}
	} else {
		auth = auths[0]
	}

	hdr.Add("Authorization", auth.GetAuthHeader())
	hdr.Add("User-agent", fmt.Sprintf("ondevice v%s", config.GetVersion()))

	url := auth.GetURL(endpoint+"/websocket", params, "wss")
	logg.Debugf("Opening websocket connection to '%s' (auth: '%s')", url, auth.GetAuthHeader())

	websocket.DefaultDialer.HandshakeTimeout = 60 * time.Second
	ws, resp, err := websocket.DefaultDialer.Dial(url, hdr)
	if err != nil {
		if resp != nil {
			if resp.StatusCode == 401 {
				return util.NewAPIError(resp.StatusCode, "API server authentication failed")
			}
			return util.NewAPIError(resp.StatusCode, "Error opening websocket: ", err)
		}
		return util.NewAPIError(util.OtherError, "Error opening websocket: ", err)
	}

	c.ws = ws
	c.MessageListerners = append(c.MessageListerners, onMessage)
	c.done = make(chan struct{})

	go c.receive()

	return nil
}

// Close -- Close the underlying WebSocket connection
func (c *Connection) Close() {
	if c.isClosed {
		return
	}

	c._onClose()
}

func (c *Connection) receive() {
	defer c._onClose()

	for {
		msgType, msg, err := c.ws.ReadMessage()
		if err != nil {
			if e, ok := err.(*websocket.CloseError); ok {
				if e.Code == 1000 {
					// normal close
				} else {
					logg.Error("Websocket closed abnormally: ", err)
				}
			} else {
				if !c.isClosed {
					logg.Errorf("read error (type: %s): %s", reflect.TypeOf(err), err)
					for _, cb := range c.ErrorListeners {
						cb(util.NewAPIError(util.OtherError, err.Error()))
					}
				} else {
					logg.Debug("Connetion.receive() interrupted by error: ", reflect.TypeOf(err), err)
				}
			}
			return
		}
		for _, cb := range c.MessageListerners {
			cb(msgType, msg)
		}
	}
}

// SendBinary -- Send binary WebSocket message
func (c *Connection) SendBinary(data []byte) {
	c.writeLock.Lock()
	c.ws.WriteMessage(websocket.BinaryMessage, data)
	c.writeLock.Unlock()
}

// SendJSON -- Send a JSON text message to the WebSocket
func (c *Connection) SendJSON(value interface{}) {
	c.writeLock.Lock()
	c.ws.WriteJSON(value)
	c.writeLock.Unlock()
}

// SendText -- send a raw text websocket messge (use SendJson instead where possible)
func (c *Connection) SendText(msg string) {
	c.writeLock.Lock()
	c.ws.WriteMessage(websocket.TextMessage, []byte(msg))
	c.writeLock.Unlock()
}

// Wait -- Wait for the connection to close
func (c *Connection) Wait() {
	<-c.done
}

func (c *Connection) _onClose() {
	if c.ws != nil { // could happen if a goroutine called us before we are connected
		c.ws.Close()
	}

	if !c.isClosed {
		c.isClosed = true
		close(c.done)
		for _, cb := range c.CloseListeners {
			cb()
		}
	}
}
