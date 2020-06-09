package config

import (
	"github.com/ondevice/ondevice/config/internal"
)

// ErrNoClientAuth -- returned by AuthConfig.GetClientAuth() if no credentials have been found
var ErrNoClientAuth = internal.ErrNoClientAuth

// ErrNoDeviceAuth -- returned by AuthConfig.GetDeviceAuth() if no credentials have been found
var ErrNoDeviceAuth = internal.ErrNoDeviceAuth

// Auth -- contains API server credentials
type Auth = internal.Auth

// AuthConfig -- loads/stores ondevice.io credentials
//
// implemented by config.internal.AuthJSON
type AuthConfig interface {
	// GetClientAuth -- returns the main client credentials
	//
	// will fail with ErrNoClientAuth if auth.json wasn't found or doesn't contain client credentials
	GetClientAuth() (Auth, error)

	GetClientAuthForDevice(devID string) (Auth, error)
	GetClientAuthForUser(username string) (Auth, error)

	// GetDeviceAuth -- returns the stored device credentials
	//
	// will fail with ErrNoDeviceAuth if auth.json wasn't found or doesn't contain device credentials
	GetDeviceAuth() (Auth, error)

	// GetDeviceKey -- returns the unique key identifying this device
	//
	// This key remains constant even if the devId changes (i.e. the device gets renamed).
	// The ondevice servers will tell us to update it if there's been a conflict (i.e. a clone has been detected)
	GetDeviceKey() string

	// IsChanged -- returns true once SetClientAuth() or SetDeviceAuth() has been called
	IsChanged() bool

	// ListClientUsers -- Returns list of authenticated client users (used for tab completion)
	ListClientUsers() []string

	// SetClientAuth -- update the client credentials
	//
	// You need to call Write() to actually update auth.json
	SetClientAuth(user string, key string)

	// SetDeviceAuth -- update device credentials
	//
	// You need to call Write() to actually update auth.json
	SetDeviceAuth(user string, key string)

	// SetDeviceKey -- update unique device key
	//
	// Returns true if the key has changed
	// You need to call Write() to actually update auth.json
	SetDeviceKey(newKey string) (changed bool)

	// updates auth.json with the credentials stored in this struct
	//
	// make sure not quickly call LoadAuth(), Set*Auth() and Write() to prevent concurrent write issues (i.e. overwrite changes by another instance)
	Write() error
}

// NewAuth -- creates new API server credentials
func NewAuth(user, authKey string) Auth {
	return internal.AuthEntry{
		UserField: user,
		KeyField:  authKey,
	}

}

// LoadAuth -- shorthand for MustLoad().LoadAuth()
func LoadAuth() AuthConfig {
	return MustLoad().LoadAuth()
}
