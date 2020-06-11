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

	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/daemon"
	"github.com/sirupsen/logrus"
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
	mux.HandleFunc("/state", rc.getStateHandler)
	mux.HandleFunc("/login", rc.postLoginHandler)
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
		logrus.Fatal("failed to parse control socket URL: ", u.String())
	}

	go c.run(proto, path)
}

// Stop -- Stops the ControlSocket
func (c *ControlSocket) Stop() error {
	var ctx, cancelFn = context.WithTimeout(context.Background(), 5*time.Second)
	var err = c.server.Shutdown(ctx)
	if err != nil {
		logrus.Error("failed to stop ControlSocket: ", err)
	} else {
		logrus.Info("stopped ControlSocket")
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
		logrus.WithError(err).Fatal("couldn't set up control socket")
	}
}

// getStateHandler -- implements GET /state
func (c *ControlSocket) getStateHandler(w http.ResponseWriter, r *http.Request) {
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

	data.Device["devId"] = config.MustLoad().GetString(config.KeyDeviceID)
	_sendJSON(w, data)
}

// postLoginHandler -- implements POST /login
func (c *ControlSocket) postLoginHandler(w http.ResponseWriter, r *http.Request) {
	var auth = config.NewAuth(r.FormValue("user"), r.FormValue("auth"))
	if auth.User() == "" || auth.Key() == "" {
		_sendError(w, http.StatusBadRequest, "expected valid user/auth")
		return
	}

	var keyInfo, err = api.GetKeyInfo(auth)
	if err != nil {
		logrus.WithError(err).Error("failed to fetch key info")
		_sendError(w, http.StatusInternalServerError, "failed to fetch key info")
	}
	if !keyInfo.IsType("device") {
		logrus.Error("not 'ondevice login' sent us credentials that aren't device credentials")
		_sendError(w, http.StatusBadRequest, "expected device credentials")
	}

	var authJSON = config.MustLoad().LoadAuth()
	authJSON.SetDeviceAuth(auth.User(), auth.Key())
	if err = authJSON.Write(); err != nil {
		logrus.WithError(err).Error("failed to update device credentials")
		_sendError(w, http.StatusInternalServerError, "failed to update device credentials")
	}
}

func _sendError(w http.ResponseWriter, statusCode int, msg string) {
	json.Marshal(struct {
		ErrorCode int
		Error     string
	}{
		ErrorCode: statusCode,
		Error:     msg,
	})
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
}

func _sendJSON(w http.ResponseWriter, data interface{}) {
	d, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).Fatal("JSON encode failed: ", data)
	}

	//logrus.Debug("Sending JSON response: ", string(d))
	w.Header().Set("Content-type", "application/json")
	io.WriteString(w, string(d))
}
