package control

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/daemon"
	"github.com/ondevice/ondevice/logg"
)

// ControlSocket instance
type ControlSocket struct {
	Daemon *daemon.DeviceSocket
}

// TODO make the socket path configurable (something in the likes of DOCKER_HOST for docker)
func getSocketPath() (string, string) {
	//	return "tcp", "localhost:1236"
	path := config.GetConfigPath("ondevice.sock")
	return "unix", path
}

// StartServer -- Start the unix domain socket server (probably won't work on Windows)
func StartServer() *ControlSocket {
	rc := ControlSocket{}
	go rc.run()
	return &rc
}

func (c *ControlSocket) run() {
	protocol, path := getSocketPath()
	if protocol == "unix" {
		os.Remove(path)
		defer os.Remove(path)
	}

	l, err := net.Listen(protocol, path)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/state", c.getState)

	err = http.Serve(l, nil)
	log.Fatal(err)
}

func (c *ControlSocket) getState(w http.ResponseWriter, req *http.Request) {
	/*	data := map[string]interface{}{
		"version": "0.0.0",
		"device": map[string]string{
			"state": "offline",
		},
	}*/

	devState := "offline"
	if c.Daemon != nil && c.Daemon.IsOnline {
		devState = "online"
	}

	data := DeviceState{
		Version: config.GetVersion(),
		Device: map[string]string{
			"state": devState,
		},
	}

	data.Device["devId"] = config.GetDeviceID()
	_sendJSON(w, data)
}

func _sendJSON(w http.ResponseWriter, data interface{}) {
	d, err := json.Marshal(data)
	if err != nil {
		logg.Fatal("JSON encode failed: ", data)
	}
	// TODO make sure we're not messing up the encoding

	logg.Debug("Sending JSON response: ", string(d))
	io.WriteString(w, string(d))
}
