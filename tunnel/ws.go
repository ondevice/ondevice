package tunnel

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/websocket"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
)

// Connection -- WebSocket connection
type Connection struct {
	ws       *websocket.Conn
	isClosed bool

	CloseListeners    []func()
	ErrorListeners    []func(err error)
	MessageListerners []func(int, []byte)

	done chan struct{}
}

// AuthenticationError -- error indicating authentication issues
type AuthenticationError struct {
	msg string
}

// OpenWebsocket -- Open a websocket connection
func OpenWebsocket(c *Connection, endpoint string, params map[string]string, onMessage func(int, []byte), auths ...api.Authentication) error {
	hdr := http.Header{}

	var auth api.Authentication
	if len(auths) == 0 {
		var err error
		if auth, err = api.CreateClientAuth(); err != nil {
			return err
		}
	} else {
		auth = auths[0]
	}

	hdr.Add("Authorization", auth.GetAuthHeader())
	hdr.Add("User-agent", fmt.Sprintf("ondevice v%s", config.GetVersion()))

	url := auth.GetURL(endpoint+"/websocket", params, "wss")
	logg.Debugf("Opening websocket connection to '%s' (auth: '%s')", url, auth.GetAuthHeader())

	ws, resp, err := websocket.DefaultDialer.Dial(url, hdr)
	if err != nil {
		if resp != nil {
			if resp.StatusCode == 401 {
				return AuthenticationError{"API server authentication failed"}
			}
			return fmt.Errorf("Error opening websocket (response code: %s): %s", resp.Status, err)
		}
		return fmt.Errorf("Error opening websocket: %s", err)
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
	c.isClosed = true
	c.ws.Close()
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
						cb(err)
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
	c.ws.WriteMessage(websocket.BinaryMessage, data)
}

// SendJSON -- Send a JSON text message to the WebSocket
func (c *Connection) SendJSON(value interface{}) {
	c.ws.WriteJSON(value)
}

// SendText -- send a raw text websocket messge (use SendJson instead where possible)
func (c *Connection) SendText(msg string) {
	c.ws.WriteMessage(websocket.TextMessage, []byte(msg))
}

// Wait -- Wait for the connection to close
func (c *Connection) Wait() {
	<-c.done
}

func (c *Connection) _onClose() {
	close(c.done)
	c.ws.Close()

	if !c.isClosed {
		c.isClosed = true
		for _, cb := range c.CloseListeners {
			cb()
		}
	}
}

func (e AuthenticationError) Error() string {
	return e.msg
}
