package rest

// KeyInfo -- API key info
type KeyInfo struct {
	Roles []string
}

// GetKeyInfo -- Returns the roles associated with the given credentials - or nil on error
func GetKeyInfo(auth Authentication) ([]string, error) {
	rc := KeyInfo{}
	err := getObject(&rc, "/keyInfo", nil, auth)
	if err != nil {
		return nil, err
	}

	return rc.Roles, nil
}
