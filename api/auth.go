package api

import (
	"strings"

	"github.com/ondevice/ondevice/config"
)

// GetClientAuthForDevice -- Returns credentials for the given devID
//
// with unqualified devIDs, this will do the same as GetClientAuth().
// But if the devID has a user prefix (and we have extra credentials for that user), it'll return those instead
func GetClientAuthForDevice(devID string) (config.Auth, error) {
	if strings.Contains(devID, ".") {
		parts := strings.SplitN(devID, ".", 2)
		if auth, err := config.GetClientAuthForUser(parts[0]); err == nil {
			return auth, nil
		}
	}

	return config.GetClientAuth()
}
