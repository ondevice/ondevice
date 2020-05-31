package internal

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
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

// AuthJSON -- marshals/unmarshals the contents of the auth.json file
type AuthJSON struct {
	Client       AuthEntry
	Device       AuthEntry
	ExtraClients []AuthEntry `json:",omitempty"`
}

// AuthEntry -- implements the Auth interface
type AuthEntry struct {
	UserField      string `json:"User"`
	KeyField       string `json:"Auth"`
	APIServerField string `json:"APIServer,omitempty"`

	// DeviceKey - stores the unique identifier key for this device
	DeviceKey string `json:",omitempty"`
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

// ReadAuth -- Read auth.json from the given file path
//
// for now, this will migrate the old-style ondevice.conf auth info into the new auth.json
func ReadAuth(path string) (AuthJSON, error) {
	var file *os.File
	var err error
	var rc AuthJSON

	if file, err = os.Open(path); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to open auth.json")
		return rc, err
	}

	var stat os.FileInfo
	if stat, err = file.Stat(); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to check auth.json file mode")
		return rc, err
	}

	if stat.Mode() != 0o600 {
		// Note: this may fail on certain platforms - TODO add special cases as they happen
		logrus.WithField("path", path).Fatalf("auth.json isn't supposed to be accessible by other users (mode: %o)", stat.Mode())
	}

	var data []byte
	if data, err = ioutil.ReadAll(file); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to read auth.json")
		return rc, err
	}

	if err = json.Unmarshal(data, &rc); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to parse auth.json")
		return rc, err
	}

	// legacy overrides (probably will be deleted in v1.0)
	if user, key := os.Getenv("ONDEVICE_USER"), os.Getenv("ONDEVICE_AUTH"); user != "" || key != "" {
		logrus.Warning("ONDEVICE_USER/ONDEVICE_AUTH overrides present. these will become obsolete soon")
		rc.Client = AuthEntry{
			UserField: os.Getenv("ONDEVICE_USER"),
			KeyField:  os.Getenv("ONDEVICE_AUTH"),
		}
		rc.Device = AuthEntry{
			UserField: os.Getenv("ONDEVICE_USER"),
			KeyField:  os.Getenv("ONDEVICE_AUTH"),
			DeviceKey: rc.Device.DeviceKey,
		}
	}

	return rc, nil
}

func init() {
	if os.Getenv("ONDEVICE_SERVER") != "" {
		_apiServer = os.Getenv("ONDEVICE_SERVER")
	}
}

var _apiServer = "https://api.ondevice.io/"
