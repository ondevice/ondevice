package daemon

import (
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
	"github.com/ondevice/ondevice/service"
	"github.com/ondevice/ondevice/tunnel"
	"github.com/ondevice/ondevice/util"
)

// ControlSocket -- REST API to control the ondevice daemon (the implementation's in the control package)
type ControlSocket interface {
	Start()
	Stop() error
}

// Daemon -- represents a device's connection to the ondevice.io API server
type Daemon struct {
	tunnel.Connection

	PIDFile    string
	ConfigFile string

	lastPing   time.Time
	wdog       *util.Watchdog
	signalChan chan os.Signal
	lock       lockFile

	IsOnline bool // TODO replace by state enum

	Control      ControlSocket
	OnConnection func(tunnelID string, service string, protocol string)
	OnError      func(error)
}

type pingMsg struct {
	Type string `json:"_type"`
	Ts   int    `json:"ts"`
}

// NewDaemon -- Create a new Daemon instance
func NewDaemon() *Daemon {
	return &Daemon{
	//		signalChan: make(chan os.Signal, 1),
	}
}

// Run -- run ondevice daemon (and return with the exit code of the command)
func (d *Daemon) Run() int {
	d.lock.Path = d.PIDFile
	if !d.lock.TryLock() {
		logg.Fatal("Couldn't acquire lock file")
		return -1
	}
	defer d.lock.Unlock()

	if d.Control != nil {
		d.Control.Start()
		defer d.Control.Stop()
	}

	//	signal.Notify(d.signalChan, syscall.SIGTERM)

	// TODO implement a sane way to stop this infinite loop (at least SIGTERM, SIGINT or maybe a unix socket call)
	retryDelay := 10 * time.Second
	for true {
		d, err := Connect()
		if err != nil {
			// only abort here if it's an authentication issue
			if err.Code() == util.AuthenticationError {
				logg.Fatal(err)
			}

			// keep retryDelay between 10 and 120sec
			if retryDelay > 120*time.Second {
				retryDelay = 120 * time.Second
			}
			if retryDelay < 10*time.Second {
				retryDelay = 10 * time.Second
			}
			// ... unless we've been rate-limited
			if err.Code() == util.TooManyRequestsError {
				retryDelay = 600 * time.Second
			}

			logg.Debug("device error: ", err)
			logg.Errorf("device error - retrying in %ds", retryDelay/time.Second)

			// sleep to avoid flooding the servers
			time.Sleep(retryDelay)

			// slowly increase retryDelay with each failed attempt
			retryDelay = time.Duration(float32(retryDelay) * 1.5)
			continue
		}

		d.Wait()

		// connection was successful -> restart after 10sec
		logg.Warning("lost device connection, reconnecting in 10s")
		retryDelay = 10
		time.Sleep(retryDelay * time.Second)
	}

	return 0
}

// Connect -- Go online
func Connect(auths ...api.Authentication) (*Daemon, util.APIError) {
	params := map[string]string{"key": config.GetDeviceKey()}
	rc := Daemon{}

	if len(auths) == 0 {
		auth, err := api.GetDeviceAuth()
		if err != nil {
			logg.Fatal("Couldn't get device auth: ", err)
		}
		auths = []api.Authentication{auth}
	}

	if err := tunnel.OpenWebsocket(&rc.Connection, "/serve", params, rc.onMessage, auths...); err != nil {
		return nil, err
	}

	rc.wdog = util.NewWatchdog(180*time.Second, rc.onTimeout)
	return &rc, nil
}

// SendConnectionError -- Send an connection error message to the API server)
func (d *Daemon) SendConnectionError(code int, msg string, tunnelID string) {
	logg.Debugf("Sending connection error: %s (code %d)", msg, code)
	data := make(map[string]interface{})
	data["_type"] = "connectError"
	data["tunnelId"] = tunnelID
	data["code"] = code
	data["msg"] = msg
	d.SendJSON(data)
}

func (d *Daemon) announce(service string, protocol string) {
	var data = map[string]string{"_type": "announce", "name": service, "protocol": protocol}
	d.SendJSON(data)
}

func (d *Daemon) onConnect(msg *map[string]interface{}) {
	clientIP := _getString(msg, "clientIp")
	clientUser := _getString(msg, "clientUser")
	protocol := _getString(msg, "protocol")
	svc := _getString(msg, "service")
	brokerURL := _getString(msg, "broker")
	tunnelID := _getString(msg, "tunnelId")

	logg.Infof("Connection request for %s:%s from user %s@%s", protocol, svc, clientUser, clientIP)

	handler := service.GetServiceHandler(svc, protocol)
	if handler == nil {
		// TODO send the error back to the API server
		logg.Error("Coudln't find protocol handler: ", protocol)
		return
	}

	service.Start(handler, tunnelID, brokerURL)
}

func (d *Daemon) onError(msg *map[string]interface{}) {
	code := _getInt(msg, "code")
	message := _getString(msg, "msg")
	var codeName = tunnel.GetErrorCodeName(int(code))

	d.IsOnline = false
	logg.Errorf("Device ERROR: %s - %s ", codeName, message)
}

func (d *Daemon) onHello(msg *map[string]interface{}) {
	logg.Debug("Got hello message: ", msg)
	var devID, key string
	if _contains(msg, "name") {
		// deprecated hello message format (for backwards compatibility) -- 2017-01-19
		devID = _getString(msg, "name")
		key = _getString(msg, "devId")
	} else {
		devID = _getString(msg, "devId")
		key = _getString(msg, "key")
	}

	logg.Infof("Connection established, online as '%s'", devID)
	d.IsOnline = true

	// update the key if changed
	if config.GetDeviceKey() != key {
		logg.Debug("Updating device key: ", key)
		config.SetValue("device", "key", key)
	}

	// update devID
	config.SetValue("device", "dev-id", devID)

	// TODO announce configured services
	d.announce("ssh", "ssh")
}

func (d *Daemon) onMessage(_type int, data []byte) {
	// got message from the API server -> reset watchdog
	d.wdog.Kick()

	if _type == websocket.BinaryMessage {
		logg.Error("Got a binary message over the device websocket: ", string(data))
		return
	}

	msg := new(map[string]interface{})

	if err := json.Unmarshal(data, msg); err != nil {
		logg.Fatalf("Malformed device message: %s", data)
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
		d.onConnect(msg)
	case "error":
		d.onError(msg)
	default:
		logg.Error("Unsupported WS message: ", data)
		break
	}
}

func (d *Daemon) onPing(msg pingMsg) {
	// quick'n'dirty way to see if we're leaking goroutines (e.g. with stray bloking reads)
	logg.Debugf("Got ping message: %+v (active goroutines: %d)", msg, runtime.NumGoroutine())
	d.lastPing = time.Now()
	resp := make(map[string]interface{}, 1)
	resp["_type"] = "pong"
	resp["ts"] = msg.Ts
	d.SendJSON(resp)
}

func (d *Daemon) onTimeout() {
	logg.Warning("Haven't heard from the API server in a while, closing connection...")
	d.IsOnline = false
	d.Close()
	d.wdog.Stop()
}

func _contains(m *map[string]interface{}, key string) bool {
	_, ok := (*m)[key]
	return ok
}

func _getInt(m *map[string]interface{}, key string) int64 {
	return (*m)[key].(int64)
}

func _getString(m *map[string]interface{}, key string) string {
	if val, ok := (*m)[key]; ok {
		var rc, ok = val.(string)
		if !ok {
			logg.Warningf("Not a string (key %s): %s", key, val)
			return ""
		}
		return rc
	}
	return ""
}
