package rest

import (
	"encoding/base64"
	"log"
	"net/url"
	"path"
	"strings"

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

func (a Authentication) getURL(endpoint string, params map[string]string) string {
	server := a.apiServer

	if server == "" {
		server = "https://api.ondevice.io/"
	}

	u, err := url.Parse(server)
	if err != nil {
		log.Fatal("URL parsing error: ", err)
	}

	if strings.HasPrefix(endpoint, "/") {
		endpoint = endpoint[1:]
	}
	u.Path = path.Join("/v1.1", endpoint)

	// parse query params
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}
	u.RawQuery = query.Encode()

	return u.String()
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
