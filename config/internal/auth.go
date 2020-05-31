package internal

import (
	"encoding/base64"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

// Auth -- contains immutable API server credentials
type Auth interface {
	// User -- user name
	User() string
	// Key -- API key for the given user
	Key() string

	// APIServer -- API server these credentials are valid for - if "", defaults to https://api.ondevice.io/
	APIServer() string

	// GetAuthHeader -- Encode the credentials into a HTTP Basic auth header
	GetAuthHeader() string

	// GetURL -- Get the full API server URL for the apiServer we store internally and the params we get
	GetURL(endpoint string, params map[string]string, scheme string) string

	// WithAPIServer -- returns a copy of this object, but with the given API server
	WithAPIServer(newServer string) Auth
}

// AuthEntry -- implements the Auth interface
type AuthEntry struct {
	UserField      string
	KeyField       string
	APIServerField string
}

// User -- returns the API user name
func (a AuthEntry) User() string { return a.UserField }

// Key -- returns the API key
func (a AuthEntry) Key() string { return a.KeyField }

// APIServer -- returns the API server URL
func (a AuthEntry) APIServer() string { return a.APIServerField }

// GetAuthHeader -- Return the value of the HTTP Basic Authentication header
func (a AuthEntry) GetAuthHeader() string {
	token := a.User() + ":" + a.Key()
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(token))
}

// GetURL -- Get the full API server URL for the apiServer we store internally and the params we get
func (a AuthEntry) GetURL(endpoint string, params map[string]string, scheme string) string {
	server := a.APIServer()

	if server == "" {
		server = _apiServer
	}

	u, err := url.Parse(server)
	if err != nil {
		logrus.WithError(err).Fatal("URL parsing error")
		return ""
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

// WithAPIServer -- returns a modified copy
func (a AuthEntry) WithAPIServer(newServer string) Auth {
	return AuthEntry{
		UserField:      a.UserField,
		KeyField:       a.KeyField,
		APIServerField: newServer,
	}
}

// NewAuth -- creates Auth credentials
func NewAuth(username, key string) Auth {
	return AuthEntry{
		UserField: username,
		KeyField:  key,
	}
}

func init() {
	if os.Getenv("ONDEVICE_SERVER") != "" {
		_apiServer = os.Getenv("ONDEVICE_SERVER")
	}
}

var _apiServer = "https://api.ondevice.io/"
