package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/tunnel"
)

// DeviceSocket -- represents a device's connection to the ondevice.io API server
type DeviceSocket struct {
	tunnel.Connection

	lastPing time.Time

	OnConnection func(tunnelID string, service string, protocol string)
	OnError      func(error)
}

type pingMsg struct {
	Type string `json:"_type"`
	Ts   int    `json:"ts"`
}

// Connect -- Go online
func Connect(auths ...api.Authentication) (*DeviceSocket, error) {
	// TODO use the new 'key' param instead of 'id'
	params := map[string]string{"id": config.GetDeviceKey()}
	rc := DeviceSocket{}

	err := tunnel.OpenWebsocket(&rc.Connection, "/serve", params, rc.onMessage, auths...)

	if err != nil {
		return nil, err
	}
	return &rc, nil
}

// SendConnectionError -- Send an connection error message to the API server)
func (d *DeviceSocket) SendConnectionError(code int, msg string, tunnelID string) {
	log.Printf("Sending connection error: %s (code %d)", msg, code)
	data := make(map[string]interface{})
	data["_type"] = "connectError"
	data["tunnelId"] = tunnelID
	data["code"] = code
	data["msg"] = msg
	d.SendJSON(data)
}

func (d *DeviceSocket) announce(service string, protocol string) {
	var data = map[string]string{"_type": "announce", "name": service, "protocol": protocol}
	d.SendJSON(data)
}

func (d *DeviceSocket) onConnect(msg *map[string]interface{}) {
	clientIP := _getString(msg, "clientIp")
	clientUser := _getString(msg, "clientUser")
	protocol := _getString(msg, "protocol")
	service := _getString(msg, "service")
	//broker := _getString(msg, "broker")
	tunnelID := _getString(msg, "tunnelId")

	// TODO support actual services
	if service != "ssh" {
		d.SendConnectionError(404, fmt.Sprintf("Service '%s' not found", service), tunnelID)
	}
	if protocol != "ssh" {
		d.SendConnectionError(400, fmt.Sprintf("Protocol mismatch (expected=%s, actual=%s)", "ssh", protocol), tunnelID)
	}

	log.Printf("Connection request for %s@%s from user %s@%s", protocol, service, clientUser, clientIP)

	d.SendConnectionError(400, "Not yet implemented", tunnelID)
}

func (d *DeviceSocket) onError(msg *map[string]interface{}) {
	code := _getInt(msg, "code")
	message := _getString(msg, "msg")
	var codeName string

	switch code {
	case 400:
		codeName = "Bad Request"
		break
	case 403:
		codeName = "Access Denied"
		break
	case 404:
		codeName = "Not Found"
		break
	}

	log.Printf("Device ERROR: %s - %s ", codeName, message)
}

func (d *DeviceSocket) onHello(msg *map[string]interface{}) {
	log.Print("Got hello message: ", msg)
	var devID, key string
	if _contains(msg, "name") {
		// deprecated hello message format (for backwards compatibility) -- 2017-01-19
		devID = _getString(msg, "name")
		key = _getString(msg, "devId")
	} else {
		devID = _getString(msg, "devId")
		key = _getString(msg, "key")
	}

	log.Printf("Connection established, online as '%s'", devID)

	// update the key if changed
	if config.GetDeviceKey() != key {
		log.Print("Updating device key: ", key)
		config.SetValue("device", "key", key)
	}

	// update devID
	config.SetValue("device", "dev-id", devID)

	// TODO announce configured services
	d.announce("ssh", "ssh")
}

func (d *DeviceSocket) onMessage(_type int, data []byte) {
	if _type == websocket.BinaryMessage {
		log.Print("Got a binary message over the device websocket: ", string(data))
		return
	}

	msg := new(map[string]interface{})

	json.Unmarshal(data, msg)

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
		log.Print("Unsupported WS message: ", data)
		break
	}
}

func (d *DeviceSocket) onPing(msg pingMsg) {
	log.Print("Got ping message: ", msg)
	d.lastPing = time.Now()
	resp := make(map[string]interface{}, 1)
	resp["_type"] = "pong"
	resp["ts"] = msg.Ts
	d.SendJSON(resp)
}

func _contains(m *map[string]interface{}, key string) bool {
	_, ok := (*m)[key]
	return ok
}

func _getInt(m *map[string]interface{}, key string) int64 {
	return (*m)[key].(int64)
}

func _getString(m *map[string]interface{}, key string) string {
	log.Printf("-- %s: %s", key, (*m)[key])
	return (*m)[key].(string)
}
