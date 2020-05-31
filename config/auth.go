package config

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func _getAuth(section string) (string, string, error) {
	if os.Getenv("ONDEVICE_USER") != "" || os.Getenv("ONDEVICE_AUTH") != "" {
		return os.Getenv("ONDEVICE_USER"), os.Getenv("ONDEVICE_AUTH"), nil
	}

	username, uerr := GetString(section, "user")
	auth, aerr := GetString(section, "auth")

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

	auth, err := GetString("client", "auth_"+username)
	if err == nil {
		return username, auth, nil
	}

	return "", "", errors.New("no client credentials found for user")
}

// GetDeviceAuth -- Get the device authentication
func GetDeviceAuth() (string, string, error) {
	return _getAuth("device")
}

// ListAuthenticatedUsers -- returns the names of users we have client auth for
func ListAuthenticatedUsers() []string {
	// TODO this is messy but will do for now - we'll improve this once we have a separate auth file

	var rc []string
	var uniqueUsers = make(map[string]bool)

	if mainUser := viper.GetString("client.user"); mainUser != "" {
		rc = append(rc, mainUser)
		uniqueUsers[strings.ToLower(mainUser)] = true
	}

	for _, k := range viper.AllKeys() {
		if strings.HasPrefix(k, "client.auth_") {
			var name = k[12:]
			var lowerName = strings.ToLower(name)
			if !uniqueUsers[lowerName] {
				rc = append(rc, k[12:])
				uniqueUsers[lowerName] = true
			}
		}
	}

	return rc
}
