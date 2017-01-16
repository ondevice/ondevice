package rest

import (
	"encoding/base64"

	"github.com/ondevice/ondevice-cli/config"
)

// Authentication -- authentication and other API options)
type Authentication struct {
	user      string
	auth      string
	apiServer string
}

func (a Authentication) getAuthHeader() string {
	token := a.user + ":" + a.auth
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(token))
}

func (a Authentication) getURL(endpoint string) string {
	server := a.apiServer
	if server == "" {
		server = "https://api.ondevice.io/"
	}
	return server + "v1.1" + endpoint
}

// CreateAuth -- Create Authentication object
func CreateAuth(user string, auth string) Authentication {
	return Authentication{
		user: user,
		auth: auth,
	}
}

// CreateClientAuth -- Create default authentication
func CreateClientAuth() (Authentication, error) {
	user, auth, err := config.GetClientAuth()
	if err != nil {
		return Authentication{}, err
	}
	return CreateAuth(user, auth), nil
}
