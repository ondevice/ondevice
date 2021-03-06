package tunnel

import (
	"time"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/util"
	"github.com/sirupsen/logrus"
)

// Connect to a service on one of your devices
func Connect(t *Tunnel, devID string, service string, protocol string, auths ...config.Auth) util.APIError {
	params := map[string]string{"dev": devID, "service": service, "protocol": protocol}

	t._initTunnel(ClientSide)
	t.TimeoutListeners = append(t.TimeoutListeners, t._sendPing)
	err := OpenWebsocket(&t.Connection, "/connect", params, t.onMessage, auths...)

	if err != nil {
		return err
	}

	// time out after 30 secs
	select {
	case err = <-t.connected:
		break
	case <-time.After(time.Second * 30):
		err = util.NewAPIError(util.OtherError, "Timeout while connecting to ", devID)
	}

	close(t.connected)
	t.connected = nil

	if err == nil {
		t.wdog = util.NewWatchdog(60*time.Second, t._onTimeout)
	}

	return err
}

func (t *Tunnel) _sendPing() {
	t.SendBinary([]byte("meta:ping:hello"))
	if t.lastPing.IsZero() {
		// ignored
	} else if t.lastPing.Add(180 * time.Second).Before(time.Now()) {
		logrus.Error("tunnel timeout, closing connection...")
		t.wdog.Stop()
		t.Close()
		return // prevent restarting the watchdog
	}
	t.wdog.Kick() // restart watchdog
}
