package tunnel

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/ondevice/ondevice/api"
)

// WSListener -- WebSocket listener
type WSListener interface {
	OnMessage(msgType int, data []byte)
}

// Connection -- WebSocket connection
type Connection struct {
	ws *websocket.Conn

	OnError   func(err error)
	onMessage func(int, []byte)
	done      chan struct{}
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

	url := auth.GetURL(endpoint+"/websocket", params, "wss")
	//log.Printf("Opening websocket connection to '%s' (auth: '%s')", url, auth.GetAuthHeader())

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
	c.onMessage = onMessage
	c.done = make(chan struct{})

	go c.receive()

	return nil
}

// Close -- Close the underlying WebSocket connection
func (c *Connection) Close() {
	c.ws.Close()
}

// OnMessage -- pass incoming WebSocket messages on to the listener function
func (c *Connection) OnMessage(msgType int, msg []byte) {
	c.onMessage(msgType, msg)
}

func (c *Connection) receive() {
	defer c.ws.Close()
	defer close(c.done)

	for {
		msgType, msg, err := c.ws.ReadMessage()
		if err != nil {
			log.Println("read error: ", err)
			if c.OnError != nil {
				c.OnError(err)
			}
			return
		}
		c.onMessage(msgType, msg)
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

func (e AuthenticationError) Error() string {
	return e.msg
}
