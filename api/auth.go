package api

import (
	"strings"

	"github.com/ondevice/ondevice/config"
)

// GetClientAuth -- Get default client authentication
func GetClientAuth() (config.Auth, error) {
	return config.GetClientAuth()
}

// GetClientAuthForDevice -- Returns credentials for the given devID
//
// with unqualified devIDs, this will do the same as GetClientAuth().
// But if the devID has a user prefix (and we have extra credentials for that user), it'll return those instead
func GetClientAuthForDevice(devID string) (config.Auth, error) {
	if strings.Contains(devID, ".") {
		parts := strings.SplitN(devID, ".", 2)
		if auth, err := config.GetClientUserAuth(parts[0]); err == nil {
			return auth, nil
		}
	}

	return GetClientAuth()
}

// GetClientAuthForUser -- Returns the client Authentication for the given user name (if available)
func GetClientAuthForUser(username string) (config.Auth, error) {
	return config.GetClientUserAuth(username)
}

// GetDeviceAuth -- Create default device authentication
func GetDeviceAuth() (config.Auth, error) {
	return config.GetDeviceAuth()
}
