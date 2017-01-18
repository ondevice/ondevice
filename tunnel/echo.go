package tunnel

import "github.com/ondevice/ondevice-cli/api"

// Echo -- open a simple echoing WebSocket connection to the API server
func Echo(onMessage func(int, []byte), auths ...api.Authentication) (*Connection, error) {
	rc := Connection{}
	err := open(&rc, "/echo", nil, onMessage, auths...)
	if err != nil {
		return nil, err
	}

	return &rc, nil
}
