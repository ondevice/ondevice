package api

import "github.com/ondevice/ondevice/config"

// KeyInfo -- API key info
type KeyInfo struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`

	// Whether these are client and/or device credentials (may be empty for disabled keys)
	Types []string `json:"types"`

	// If set, use this as auth key (instead of the one entered by the user)
	// This allows users to use their actual account password for `ondevice login`.
	// The server will transparently create a 'full' auth key and return its auth here
	Key string `json:"key"`

	// If set, show these to the user (may warn them about disabled keys, etc.)
	Messages []string `json:"messages"`
}

// HasPermission -- Checks if the auth key has the requested permission
func (i KeyInfo) HasPermission(permission string) bool {
	for _, perm := range i.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// IsType -- Checks if the AuthKey is of the specified type ('client' or 'device')
func (i KeyInfo) IsType(typeName string) bool {
	for _, t := range i.Types {
		if t == typeName {
			return true
		}
	}
	return false
}

// GetKeyInfo -- Returns the role and permissions associated with the given credentials
func GetKeyInfo(auth config.Auth) (KeyInfo, error) {
	var rc KeyInfo
	err := getObject(&rc, "/keyInfo", nil, auth)
	if err != nil {
		return rc, err
	}

	return rc, nil
}
