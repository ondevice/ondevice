package api

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

// GetAuthHeader -- Return the value of the HTTP Basic Authentication header
func (a Authentication) GetAuthHeader() string {
	token := a.user + ":" + a.auth
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(token))
}

// GetURL -- Get the full API server URL for the apiServer we store internally and the params we get
func (a Authentication) GetURL(endpoint string, params map[string]string, scheme string) string {
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

	u.Scheme = scheme
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