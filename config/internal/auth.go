package internal

// Auth -- contains immutable API server credentials
type Auth interface {
	// User name
	User() string
	// API key for the given user
	Key() string

	// API server these credentials are valid for - if "", defaults to https://api.ondevice.io/
	APIServer() string
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

// NewAuth -- creates Auth credentials
func NewAuth(username, key, apiServer string) Auth {
	return AuthEntry{
		UserField:      username,
		KeyField:       key,
		APIServerField: apiServer,
	}
}
