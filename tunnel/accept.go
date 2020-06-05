package tunnel

import (
	"fmt"
	"time"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/util"
	"github.com/sirupsen/logrus"
)

// Accept -- Accept an incoming tunnel connection
// Note: blocks until the tunnel has been established, so do this in a goroutine
func Accept(t *Tunnel, tunnelID string, relayURL string, auths ...config.Auth) error {
	var cfg, err = config.Load()
	if err != nil {
		return err
	}
	deviceKey := cfg.LoadAuth().GetDeviceKey()
	params := map[string]string{"key": deviceKey, "tunnel": tunnelID}

	// get authentication
	auth, err := config.LoadAuth().GetDeviceAuth()
	if len(auths) > 0 {
		auth = auths[0]
	} else if err != nil {
		return err
	}

	// set relayURL
	auth = auth.WithAPIServer(relayURL)

	t._initTunnel(DeviceSide)
	t.CloseListeners = append(t.CloseListeners, t._onClose)
	t.TimeoutListeners = append(t.TimeoutListeners, t._pingTimeout)
	err = OpenWebsocket(&t.Connection, "/accept", params, t.onMessage, auth)

	if err != nil {
		return err
	}

	// time out after 30 secs
	select {
	case err = <-t.connected:
		break
	case <-time.After(time.Second * 30):
		err = fmt.Errorf("Timeout while accepting tunnel %s", tunnelID)
	}

	close(t.connected)
	t.connected = nil

	// init watchdog
	if err == nil {
		t.wdog = util.NewWatchdog(180*time.Second, t._onTimeout)
	}

	return err
}

func (t *Tunnel) _pingTimeout() {
	if !t.IsClosed() {
		logrus.Error("tunnel timeout, closing connection")
		t.Close()
	}
}
