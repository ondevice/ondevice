package api

import (
	"encoding/base64"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
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
		server = _apiServer
	}

	u, err := url.Parse(server)
	if err != nil {
		logg.Fatal("URL parsing error: ", err)
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

	// use ws:// if the API server URL was http:// (for local testing)
	if u.Scheme == "http" {
		// server URL was using http:// (for local testing) -> use http:// and ws:// (instead of their secure counterparts)
		if scheme == "wss" {
			u.Scheme = "ws"
		} else if scheme == "https" {
			u.Scheme = "http"
		} else {
			u.Scheme = scheme
		}
	} else if u.Scheme == "ws" && scheme == "wss" {
		u.Scheme = "ws"
	} else {
		u.Scheme = scheme
	}

	return u.String()
}

// SetServerURL -- Set the API server's URL (used by GetURL(), necessary to use the correct API server in tunnel.Accept())
func (a *Authentication) SetServerURL(apiServer string) {
	a.apiServer = apiServer
}

// NewAuth -- Create Authentication object
func NewAuth(user string, auth string) Authentication {
	return Authentication{
		user: user,
		auth: auth,
	}
}

// CreateClientAuth -- Get default client authentication
func CreateClientAuth() (Authentication, error) {
	user, auth, err := config.GetClientAuth()
	if err != nil {
		return Authentication{}, err
	}
	return NewAuth(user, auth), nil
}

// CreateDeviceAuth -- Create default device authentication
func CreateDeviceAuth() (Authentication, error) {
	user, auth, err := config.GetDeviceAuth()
	if err != nil {
		return Authentication{}, err
	}
	return NewAuth(user, auth), nil
}

// GetClientAuthForDevice -- Returns an Authentication struct for the given devID
//
// with unqualified devIDs, this will do the same as CreateClientAuth().
// But if the devID has a user prefix (and we have extra credentials for that user), it'll return those instead
func GetClientAuthForDevice(devID string) (Authentication, error) {
	if strings.Contains(devID, ".") {
		parts := strings.SplitN(devID, ".", 2)
		if user, pwd, err := config.GetClientUserAuth(parts[0]); err == nil {
			return NewAuth(user, pwd), nil
		}
	}

	return CreateClientAuth()
}

// GetClientAuthForUser -- Returns the client Authentication for the given user name (if available)
func GetClientAuthForUser(username string) (Authentication, error) {
	var user, key, err = config.GetClientUserAuth(username)
	var rc Authentication

	if err != nil {
		return rc, err
	}
	rc = NewAuth(user, key)
	return rc, nil
}

func init() {
	if os.Getenv("ONDEVICE_SERVER") != "" {
		_apiServer = os.Getenv("ONDEVICE_SERVER")
	}
}

var _apiServer = "https://api.ondevice.io/"
