package tunnel

import (
	"fmt"
	"time"

	"github.com/ondevice/ondevice/api"
)

// Connect to a service on one of your devices
func Connect(devID string, service string, protocol string, auths ...api.Authentication) (*Tunnel, error) {
	params := map[string]string{"dev": devID, "service": service, "protocol": protocol}
	rc := Tunnel{}

	rc.connected = make(chan error)
	err := OpenWebsocket(&rc.Connection, "/connect", params, rc.onMessage, auths...)

	if err != nil {
		return nil, err
	}

	// time out after 30 secs
	go func() {
		time.Sleep(30 * time.Second)
		rc.connected <- fmt.Errorf("Timeout while connecting to %s", devID)
	}()

	err = <-rc.connected
	if err != nil {
		return nil, err
	}

	close(rc.connected)
	rc.connected = nil

	return &rc, nil
}
