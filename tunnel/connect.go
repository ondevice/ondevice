package tunnel

import (
	"fmt"
	"log"
	"time"

	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/util"
)

// Connect to a service on one of your devices
func Connect(t *Tunnel, devID string, service string, protocol string, auths ...api.Authentication) error {
	params := map[string]string{"dev": devID, "service": service, "protocol": protocol}

	t.connected = make(chan error)
	t.OnTimeout = t._sendPing
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

	if err == nil {
		t.wdog = util.NewWatchdog(60*time.Second, t.OnTimeout)
	}

	return err
}

func (t *Tunnel) _sendPing() {
	log.Print("~~sendPing~~")
	t.SendBinary([]byte("meta:ping:hell:no"))
	if t.lastPing.IsZero() {
		// ignored
	} else if t.lastPing.Add(180 * time.Second).Before(time.Now()) {
		log.Print("tunnel timeout, closing connection...")
		t.wdog.Stop()
		t.Close()
		return // prevent restarting the watchdog
	}
	t.wdog.Kick() // restart watchdog
}
