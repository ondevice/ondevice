package internal

// Auth -- contains immutable API server credentials
type Auth interface {
	// User -- user name
	User() string
	// Key -- API key for the given user
	Key() string

	// APIServer -- API server these credentials are valid for - if "", defaults to https://api.ondevice.io/
	APIServer() string

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

// WithAPIServer -- returns a modified copy
func (a AuthEntry) WithAPIServer(newServer string) Auth {
	return AuthEntry{
		UserField:      a.UserField,
		KeyField:       a.KeyField,
		APIServerField: newServer,
	}
}

// NewAuth -- creates Auth credentials
func NewAuth(username, key, apiServer string) Auth {
	return AuthEntry{
		UserField:      username,
		KeyField:       key,
		APIServerField: apiServer,
	}
}
