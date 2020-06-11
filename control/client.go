package control

import (
	"net/url"

	"github.com/ondevice/ondevice/config"
	"github.com/sirupsen/logrus"
)

// GetState -- Query the device state over the unix socket
func GetState() (DeviceState, error) {
	rc := DeviceState{}
	err := request{endpoint: "/state"}.Get().ReadJSON(&rc)
	return rc, err
}

// Login -- Send device login credentials to the ondevice daemon
//
// Note: the daemon and client might share config/auth files. In that case it's important that they don't
// concurrently try to update the auth.json -> call this AFTER you've called AuthConfig.Write()
// (or alternatively: before even calling Config.LoadAuth())
func Login(auth config.Auth) error {
	var form = make(url.Values)

	form.Set("user", auth.User())
	form.Set("key", auth.Key())

	var resp = request{endpoint: "/login"}.PostForm(form)
	if err := resp.Error(); err != nil {
		logrus.WithError(err).Error("failed to send login credentials to daemon")
	}

	return resp.Error()
}
