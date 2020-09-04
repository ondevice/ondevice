package tunnel

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/looplab/fsm"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/util"
	"github.com/sirupsen/logrus"
)

const stateInitial = "initial"
const stateConnecting = "connecting"
const stateOpen = "open"
const stateClosed = "closed"

const evConnect = "connect"
const evConnected = "connected"
const evError = "error"
const evClose = "close"

// Connection -- WebSocket connection
//
// internal state machine:
// - initial (OpenWebsocket() sets the state to 'connecting' pretty soon after creating the state machine, so this state isn't all too relevant)
// - connecting: event: "connect", from: initial
// - open: event: "connected", from: connecting
// - closed:
//	events: "close", "error"
//	from: initial, connecting, open
type Connection struct {
	ws           *websocket.Conn
	stateMachine *fsm.FSM

	CloseListeners    []func()
	ErrorListeners    []func(err util.APIError)
	MessageListerners []func(int, []byte)

	writeLock sync.Mutex
	done      chan struct{}
}

// OpenWebsocket -- Open a websocket connection
func OpenWebsocket(c *Connection, endpoint string, params map[string]string, onMessage func(int, []byte), auths ...config.Auth) util.APIError {
	if c.stateMachine != nil {
		// TODO this is not thread safe
		panic("OpenWebsocket() called twice on a single Connection!")
	}
	c.stateMachine = fsm.NewFSM(stateInitial, fsm.Events{
		{Name: "connect", Src: []string{stateInitial}, Dst: stateConnecting},
		{Name: evConnected, Src: []string{stateConnecting}, Dst: stateOpen},
		{Name: evError, Src: []string{stateInitial, stateConnecting, stateOpen}, Dst: stateClosed},
		{Name: evClose, Src: []string{stateInitial, stateConnecting, stateOpen}, Dst: stateClosed},
	}, fsm.Callbacks{
		"after_" + evError:     c._onError,
		"enter_" + stateClosed: c._onClose,
		"enter_state":          c._onStateChange,
	})
	c.done = make(chan struct{})

	hdr := http.Header{}

	var auth config.Auth
	if len(auths) == 0 {
		var err error
		if auth, err = config.LoadAuth().GetClientAuth(); err != nil {
			return util.NewAPIError(util.OtherError, err.Error())
		}
	} else {
		auth = auths[0]
	}

	hdr.Add("Authorization", auth.GetAuthHeader())
	hdr.Add("User-agent", fmt.Sprintf("ondevice v%s", config.GetVersion()))

	url := auth.GetURL(endpoint+"/websocket", params, "wss")
	logrus.Debugf("opening websocket connection to '%s' (auth: '%s')", url, auth.GetAuthHeader())

	c.stateMachine.Event(evConnect)
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

	go c.receive()

	return nil
}

// Close -- Close the underlying WebSocket connection
func (c *Connection) Close() {
	if err := c.stateMachine.Event(evClose); err != nil {
		if _, ok := err.(fsm.NoTransitionError); ok {
			// already in closed state
		} else {
			logrus.WithError(err).Error("failed to close connection")
		}
	}
}

// IsClosed -- Returns true for closed connections (either being closed normally or due to an error/timeout)
func (c *Connection) IsClosed() bool {
	if s := c.stateMachine; s != nil {
		return s.Is(stateClosed)
	}
	return false
}

func (c *Connection) receive() {
	defer c.Close()

	for {
		msgType, msg, err := c.ws.ReadMessage()
		if err != nil {
			if e, ok := err.(*websocket.CloseError); ok {
				if e.Code == 1000 {
					// normal close
				} else {
					logrus.WithError(err).Error("websocket closed abnormally")
				}
			} else {
				if !c.IsClosed() {
					logrus.WithError(err).Errorf("read error (type: %s)", reflect.TypeOf(err))
					c._error(util.NewAPIError(util.OtherError, err.Error()))
				} else {
					logrus.WithError(err).Debug("connetion.receive() interrupted by error: ", reflect.TypeOf(err))
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
func (c *Connection) SendBinary(data []byte) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	return c.ws.WriteMessage(websocket.BinaryMessage, data)
}

// SendJSON -- Send a JSON text message to the WebSocket
func (c *Connection) SendJSON(value interface{}) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	return c.ws.WriteJSON(value)
}

// SendText -- send a raw text websocket messge (use SendJson instead where possible)
func (c *Connection) SendText(msg string) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	return c.ws.WriteMessage(websocket.TextMessage, []byte(msg))
}

// Wait -- Wait for the connection to close
func (c *Connection) Wait() {
	<-c.done
}

// _error -- Puts the connection into the 'error' state and closes it
func (c *Connection) _error(err util.APIError) {
	c.stateMachine.Event(evError, err)
}

func (c *Connection) _onClose(ev *fsm.Event) {
	if c.ws != nil { // could be nil if a goroutine called us before we are connected
		c.ws.Close()
	}

	close(c.done)
	go func() { // do this asynchronously (we could end up in a deadlock (if one of the callbacks calls Close() again))
		for _, cb := range c.CloseListeners {
			cb()
		}
	}()
}

func (c *Connection) _onError(ev *fsm.Event) {
	err, ok := ev.Args[0].(util.APIError)
	if !ok {
		panic("Connection._onError() expects an APIError parameter!")
	}
	for _, cb := range c.ErrorListeners {
		cb(err)
	}
}

func (c *Connection) _onStateChange(ev *fsm.Event) {
	logrus.Debug("connection state changed: ", ev.Src, " -> ", ev.Dst)
}
