package tunnel

import "github.com/ondevice/ondevice/api"

// Connect to a service on one of your devices
func Connect(devID string, service string, protocol string, auths ...api.Authentication) (*Tunnel, error) {
	params := map[string]string{"dev": devID, "service": service, "protocol": protocol}
	rc := Tunnel{}
	rc.sendLock.Lock()

	err := OpenWebsocket(&rc.Connection, "/connect", params, rc.onMessage, auths...)

	if err != nil {
		return nil, err
	}
	return &rc, nil
}
