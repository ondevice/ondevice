package tunnel

import (
	"fmt"
	"time"

	"github.com/ondevice/ondevice/api"
)

// Connect to a service on one of your devices
func Connect(t *Tunnel, devID string, service string, protocol string, auths ...api.Authentication) error {
	params := map[string]string{"dev": devID, "service": service, "protocol": protocol}

	t.connected = make(chan error)
	err := OpenWebsocket(&t.Connection, "/connect", params, t.onMessage, auths...)

	if err != nil {
		return err
	}

	// time out after 30 secs
	select {
	case err = <-t.connected:
		break
	case <-time.After(time.Second * 30):
		err = fmt.Errorf("Timeout while connecting to %s", devID)
	}

	close(t.connected)
	t.connected = nil

	return err
}
