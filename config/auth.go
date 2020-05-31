package config

import (
	"errors"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func _getAuth(section string) (string, string, error) {
	if os.Getenv("ONDEVICE_USER") != "" || os.Getenv("ONDEVICE_AUTH") != "" {
		return os.Getenv("ONDEVICE_USER"), os.Getenv("ONDEVICE_AUTH"), nil
	}

	var cfg, err = Read()
	if err != nil {
		return "", "", err
	}

	var username, auth string
	if username, err = cfg.GetString(section, "user"); err != nil {
		return "", "", err
	}
	if auth, err = cfg.GetString(section, "auth"); err != nil {
		return "", "", err
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

	var cfg, err = Read()
	if err != nil {
		return "", "", errors.New("failed to read ondevice.conf")
	}

	var auth string
	if auth, err = cfg.GetString("client", "auth_"+username); err == nil {
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

	var cfg, err = Read()
	if err != nil {
		logrus.WithError(err).Fatal("failed to fetch configuration")
	}

	if mainUser, err := cfg.GetString("client", "user"); err != nil && mainUser != "" {
		rc = append(rc, mainUser)
		uniqueUsers[strings.ToLower(mainUser)] = true
	}

	for k := range cfg.AllValues() {
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
