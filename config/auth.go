package config

import (
	"os"

	"github.com/ondevice/ondevice/config/internal"
)

// Auth -- contains API server credentials
type Auth = internal.Auth

// NewAuth -- creates new API server credentials
func NewAuth(user, auth string) Auth {
	return internal.NewAuth(user, auth)
}

func _getAuth(section string) (Auth, error) {
	if os.Getenv("ONDEVICE_USER") != "" || os.Getenv("ONDEVICE_AUTH") != "" {
		return internal.NewAuth(os.Getenv("ONDEVICE_USER"), os.Getenv("ONDEVICE_AUTH")), nil
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

	return internal.NewAuth(username, auth), nil
}

// LoadAuth -- shorthand for MustLoad().LoadAuth()
func LoadAuth() internal.AuthJSON {
	return MustLoad().LoadAuth()
}
