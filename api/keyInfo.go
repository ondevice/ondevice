package api

// KeyInfo -- API key info
type KeyInfo struct {
	Role        string
	Permissions []string
}

// HasPermission -- Checks if the auth key has the requested permission
func (i KeyInfo) HasPermission(permission string) bool {
	for _, perm := range i.Permissions {
		if perm == string(permission) {
			return true
		}
	}
	return false
}

// GetKeyInfo -- Returns the role and permissions associated with the given credentials
func GetKeyInfo(auth Authentication) (KeyInfo, error) {
	var rc KeyInfo
	err := getObject(&rc, "/keyInfo", nil, auth)
	if err != nil {
		return rc, err
	}

	return rc, nil
}
