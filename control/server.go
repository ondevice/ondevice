package control

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/daemon"
	"github.com/ondevice/ondevice/logg"
)

// ControlSocket instance
type ControlSocket struct {
	Daemon *daemon.Daemon

	URL    url.URL
	server http.Server
}

// NewSocket -- Creates a new ControlSocket instance
func NewSocket(d *daemon.Daemon, u url.URL) *ControlSocket {
	var rc = ControlSocket{
		Daemon: d,
		URL:    u,
	}

	var mux = new(http.ServeMux)
	mux.HandleFunc("/state", rc.getState)
	rc.server.Handler = mux

	return &rc
}

// Start -- Starts the ControlSocket (parses the URL )
func (c *ControlSocket) Start() {
	var proto, path string
	var u = c.URL

	// TODO move me to NewSocket()
	if u.Scheme == "unix" || u.Scheme == "" {
		proto = "unix"
		path = u.Path
	} else if u.Scheme == "http" {
		proto = "tcp"
		path = u.Host
	} else {
		logg.Fatal("Failed to parse control socket URL: ", u.String())
	}

	go c.run(proto, path)
}

// Stop -- Stops the ControlSocket
func (c *ControlSocket) Stop() error {
	var ctx, cancelFn = context.WithTimeout(context.Background(), 5*time.Second)
	var err = c.server.Shutdown(ctx)
	if err != nil {
		logg.Error("Failed to stop ControlSocket: ", err)
	} else {
		logg.Info("Stopped ControlSocket")
	}

	cancelFn()
	return err
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

	if protocol == "unix" {
		os.Chmod(path, 0664)
	}

	err = c.server.Serve(l)
	if err != nil && err != http.ErrServerClosed {
		logg.Fatal("Couldn't set up control socket: ", err)
	}
}

func (c *ControlSocket) getState(w http.ResponseWriter, req *http.Request) {
	devState := "offline"
	if c.Daemon != nil && c.Daemon.IsOnline() {
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
