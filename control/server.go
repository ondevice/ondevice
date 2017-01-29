package control

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/daemon"
	"github.com/ondevice/ondevice/logg"
)

// ControlSocket instance
type ControlSocket struct {
	Daemon *daemon.DeviceSocket
}

// StartServer -- Start the unix domain socket server (probably won't work on Windows)
func StartServer(u url.URL) *ControlSocket {
	var proto, path string
	if u.Scheme == "unix" || u.Scheme == "" {
		proto = "unix"
		path = u.Path
	} else if u.Scheme == "http" {
		proto = "tcp"
		path = u.Host
	}

	rc := ControlSocket{}
	go rc.run(proto, path)
	return &rc
}

func (c *ControlSocket) run(protocol string, path string) {
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
