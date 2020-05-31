package tunnel

import "github.com/ondevice/ondevice/config"

// Echo -- open a simple echoing WebSocket connection to the API server
func Echo(onMessage func(int, []byte), auths ...config.Auth) (*Connection, error) {
	rc := Connection{}
	err := OpenWebsocket(&rc, "/echo", nil, onMessage, auths...)
	if err != nil {
		return nil, err
	}

	return &rc, nil
}
