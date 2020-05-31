package api

import (
	"strings"

	"github.com/ondevice/ondevice/config"
)

// Authentication -- authentication and other API options)
type Authentication struct {
	auth config.Auth
}

// GetAuthHeader -- Return the value of the HTTP Basic Authentication header
func (a Authentication) GetAuthHeader() string { return a.auth.GetAuthHeader() }

// GetURL -- Get the full API server URL for the apiServer we store internally and the params we get
func (a Authentication) GetURL(endpoint string, params map[string]string, scheme string) string {
	return a.auth.GetURL(endpoint, params, scheme)
}

// SetServerURL -- Set the API server's URL (used by GetURL(), necessary to use the correct API server in tunnel.Accept())
func (a *Authentication) SetServerURL(apiServer string) {
	a.auth = a.auth.WithAPIServer(apiServer)
}

// NewAuth -- Create Authentication object
func NewAuth(auth config.Auth) Authentication {
	return Authentication{
		auth: auth,
	}
}

// GetClientAuth -- Get default client authentication
func GetClientAuth() (Authentication, error) {
	auth, err := config.GetClientAuth()
	if err != nil {
		return Authentication{}, err
	}
	return NewAuth(auth), nil
}

// GetClientAuthForDevice -- Returns an Authentication struct for the given devID
//
// with unqualified devIDs, this will do the same as GetClientAuth().
// But if the devID has a user prefix (and we have extra credentials for that user), it'll return those instead
func GetClientAuthForDevice(devID string) (Authentication, error) {
	if strings.Contains(devID, ".") {
		parts := strings.SplitN(devID, ".", 2)
		if auth, err := config.GetClientUserAuth(parts[0]); err == nil {
			return NewAuth(auth), nil
		}
	}

	return GetClientAuth()
}

// GetClientAuthForUser -- Returns the client Authentication for the given user name (if available)
func GetClientAuthForUser(username string) (Authentication, error) {
	var auth, err = config.GetClientUserAuth(username)
	var rc Authentication

	if err != nil {
		return rc, err
	}
	rc = NewAuth(auth)
	return rc, nil
}

// GetDeviceAuth -- Create default device authentication
func GetDeviceAuth() (Authentication, error) {
	auth, err := config.GetDeviceAuth()
	if err != nil {
		return Authentication{}, err
	}
	return NewAuth(auth), nil
}
