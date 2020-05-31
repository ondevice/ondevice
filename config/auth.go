package config

import (
	"errors"
	"os"
	"strings"

	"github.com/ondevice/ondevice/config/internal"
	"github.com/sirupsen/logrus"
)

// Auth -- contains API server credentials
type Auth = internal.Auth

// NewAuth -- creates new API server credentials
func NewAuth(user, auth, apiServer string) Auth {
	return internal.NewAuth(user, auth, apiServer)
}

func _getAuth(section string) (Auth, error) {
	if os.Getenv("ONDEVICE_USER") != "" || os.Getenv("ONDEVICE_AUTH") != "" {
		return internal.NewAuth(os.Getenv("ONDEVICE_USER"), os.Getenv("ONDEVICE_AUTH"), ""), nil
	}

	var cfg, err = Read()
	if err != nil {
		return nil, err
	}

	var username, auth string
	if username, err = cfg.GetString(section, "user"); err != nil {
		return nil, err
	}
	if auth, err = cfg.GetString(section, "auth"); err != nil {
		return nil, err
	}

	return internal.NewAuth(username, auth, ""), nil
}

// GetClientAuth -- Get the default client authentication information
func GetClientAuth() (Auth, error) {
	return _getAuth("client")
}

// GetClientAuthForUser -- get the authentication credentials for a specific client user
func GetClientAuthForUser(username string) (Auth, error) {
	var auth, err = GetClientAuth()
	if err == nil && strings.ToLower(auth.User()) == strings.ToLower(username) {
		return auth, nil
	}

	var cfg Config
	if cfg, err = Read(); err != nil {
		return nil, errors.New("failed to read ondevice.conf")
	}

	var key string
	if key, err = cfg.GetString("client", "auth_"+username); err == nil {
		return internal.NewAuth(username, key, ""), nil
	}

	return nil, errors.New("no client credentials found for user")
}

// GetDeviceAuth -- Get the device authentication
func GetDeviceAuth() (Auth, error) {
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
