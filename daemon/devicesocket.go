package daemon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/service"
	"github.com/ondevice/ondevice/tunnel"
	"github.com/ondevice/ondevice/util"
	"github.com/sirupsen/logrus"
)

// deviceSocket -- Manages a single connection to the API servers
type deviceSocket struct {
	tunnel.Connection

	IsOnline bool

	wdog          *util.Watchdog
	lastPing      time.Time
	activeTunnels *sync.WaitGroup
}

func (d *deviceSocket) announce(service string, protocol string) {
	var data = map[string]string{"_type": "announce", "name": service, "protocol": protocol}
	d.SendJSON(data)
}

// connect -- Go online
func (d *deviceSocket) connect(auths ...config.Auth) util.APIError {
	var cfg, err = config.Read()
	if err != nil {
		return util.NewAPIError(500, err.Error())
	}

	params := map[string]string{"key": cfg.GetDeviceKey()}

	if len(auths) == 0 {
		auth, err := config.GetDeviceAuth()
		if err != nil {
			logrus.WithError(err).Fatal("couldn't get device auth")
		}
		auths = []config.Auth{auth}
	}

	if err := tunnel.OpenWebsocket(&d.Connection, "/serve", params, d.onMessage, auths...); err != nil {
		return err
	}

	d.wdog = util.NewWatchdog(180*time.Second, d.onTimeout)
	return nil
}

func (d *deviceSocket) onConnect(msg *map[string]interface{}) {
	clientIP := _getString(msg, "clientIp")
	clientUser := _getString(msg, "clientUser")
	protocol := _getString(msg, "protocol")
	svc := _getString(msg, "service")
	brokerURL := _getString(msg, "broker")
	tunnelID := _getString(msg, "tunnelId")

	logrus.Infof("connection request for %s:%s from user %s@%s", protocol, svc, clientUser, clientIP)

	handler := service.GetServiceHandler(svc, protocol)
	if handler == nil {
		d.SendConnectionError(http.StatusNotFound, fmt.Sprintf("Couldn't find service: '%s'", svc), tunnelID)
		logrus.Error("coudln't find protocol handler: ", protocol)
		return
	}

	d.activeTunnels.Add(1)
	service.Run(handler, tunnelID, brokerURL)
	d.activeTunnels.Done()
}

func (d *deviceSocket) onError(msg *map[string]interface{}) {
	code := _getInt(msg, "code")
	message := _getString(msg, "msg")
	var codeName = tunnel.GetErrorCodeName(int(code))

	d.IsOnline = false
	logrus.Errorf("device ERROR: %s - %s ", codeName, message)
}

func (d *deviceSocket) onHello(msg *map[string]interface{}) {
	logrus.Debug("got hello message: ", msg)
	var devID, key string
	if _contains(msg, "name") {
		// deprecated hello message format (for backwards compatibility) -- 2017-01-19
		devID = _getString(msg, "name")
		key = _getString(msg, "devId")
	} else {
		devID = _getString(msg, "devId")
		key = _getString(msg, "key")
	}

	logrus.Infof("connection established, online as '%s'", devID)
	d.IsOnline = true

	var cfg, err = config.Read()
	if err != nil {
		logrus.WithError(err).Fatal("failed to read ondevice.conf")
	}

	// update the key if changed
	if cfg.GetDeviceKey() != key {
		logrus.Debug("updating device key: ", key)
		cfg.SetValue("device", "key", key)
	}

	// update devID
	if cfg.GetDeviceID() != devID {
		cfg.SetValue("device", "dev-id", devID)
	}

	if cfg.IsChanged() {
		cfg.Write()
	}

	// TODO announce configured services
	d.announce("ssh", "ssh")
}

func (d *deviceSocket) onMessage(_type int, data []byte) {
	// got message from the API server -> reset watchdog
	d.wdog.Kick()

	if _type == websocket.BinaryMessage {
		logrus.Error("got a binary message over the device websocket: ", string(data))
		return
	}

	msg := new(map[string]interface{})

	if err := json.Unmarshal(data, msg); err != nil {
		logrus.Fatalf("malformed device message: %s", data)
	}

	msgType := _getString(msg, "_type")
	switch msgType {
	case "hello":
		d.onHello(msg)
		break
	case "ping":
		var ping pingMsg
		json.Unmarshal(data, &ping)
		d.onPing(ping)
		break
	case "connect":
		go d.onConnect(msg)
	case "error":
		d.onError(msg)
	default:
		logrus.Error("unsupported WebSocket message: ", data)
		break
	}
}

func (d *deviceSocket) onPing(msg pingMsg) {
	// quick'n'dirty way to see if we're leaking goroutines (e.g. with stray bloking reads)
	logrus.Debugf("got ping message: %+v (active goroutines: %d)", msg, runtime.NumGoroutine())
	d.lastPing = time.Now()
	resp := make(map[string]interface{}, 1)
	resp["_type"] = "pong"
	resp["ts"] = msg.Ts
	d.SendJSON(resp)
}

func (d *deviceSocket) onTimeout() {
	logrus.Warning("haven't heard from the API server in a while, closing connection...")
	d.IsOnline = false
	d.Close()
	d.wdog.Stop()
}

// SendConnectionError -- Send an connection error message to the API server)
func (d *deviceSocket) SendConnectionError(code int, msg string, tunnelID string) {
	logrus.Debugf("sending connection error: %s (code %d)", msg, code)
	data := make(map[string]interface{})
	data["_type"] = "connectError"
	data["tunnelId"] = tunnelID
	data["code"] = code
	data["msg"] = msg
	d.SendJSON(data)
}
