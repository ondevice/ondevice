package config

import (
	"errors"
	"os"
)

func _getAuth(section string) (string, string, error) {
	if os.Getenv("ONDEVICE_USER") != "" || os.Getenv("ONDEVICE_AUTH") != "" {
		return os.Getenv("ONDEVICE_USER"), os.Getenv("ONDEVICE_AUTH"), nil
	}

	username, uerr := GetValue(section, "user")
	auth, aerr := GetValue(section, "auth")

	if uerr != nil {
		return "", "", uerr
	}
	if aerr != nil {
		return "", "", aerr
	}

	return username, auth, nil
}

// GetClientAuth -- Get the default client authentication information
func GetClientAuth() (string, string, error) {
	return _getAuth("client")
}

// GetClientUserAuth -- get the authentication credentials for a specific client user
func GetClientUserAuth(username string) (string, string, error) {
	defaultU, defaultA, defaultE := GetClientAuth()
	if defaultE == nil && defaultU == username {
		return defaultU, defaultA, nil
	}

	auth, err := GetValue("client", "auth_"+username)
	if err == nil {
		return username, auth, nil
	}

	return "", "", errors.New("no client credentials found for user")
}

// GetDeviceAuth -- Get the device authentication
func GetDeviceAuth() (string, string, error) {
	return _getAuth("device")
}
