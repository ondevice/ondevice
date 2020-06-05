package config

import (
	"github.com/ondevice/ondevice/config/internal"
)

// Auth -- contains API server credentials
type Auth = internal.Auth

// NewAuth -- creates new API server credentials
func NewAuth(user, auth string) Auth {
	return internal.NewAuth(user, auth)
}

// LoadAuth -- shorthand for MustLoad().LoadAuth()
func LoadAuth() internal.AuthJSON {
	return MustLoad().LoadAuth()
}
