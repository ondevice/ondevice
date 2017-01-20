package tunnel

import (
	"fmt"
	"time"

	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
)

// Accept -- Accept an incoming tunnel connection
// Note: blocks until the tunnel has been established, so do this in a goroutine
func Accept(t *Tunnel, tunnelID string, brokerURL string, auths ...api.Authentication) error {
	deviceKey := config.GetDeviceKey()
	params := map[string]string{"key": deviceKey, "tunnel": tunnelID}

	// get authentication
	auth, err := api.CreateDeviceAuth()
	if len(auths) > 0 {
		auth = auths[0]
	} else if err != nil {
		return err
	}

	// set brokerURL
	auth.SetServerURL(brokerURL)

	t.connected = make(chan error)
	err = OpenWebsocket(&t.Connection, "/accept", params, t.onMessage, auth)

	if err != nil {
		return err
	}

	// time out after 30 secs
	go func() {
		time.Sleep(30 * time.Second)
		t.connected <- fmt.Errorf("Timeout while accepting tunnel %s", tunnelID)
	}()

	err = <-t.connected
	if err != nil {
		return err
	}

	close(t.connected)
	t.connected = nil

	return nil
}
