package config

// GetDeviceID -- Returns the devId if available (otherwise returns an empty string)
func GetDeviceID() string {
	rc, err := GetValue("device", "dev-id")
	if err != nil {
		return ""
	}
	return rc
}

// GetDeviceKey -- Returns the device's key (or an empty string if not defined)
func GetDeviceKey() string {
	rc, err := GetValue("device", "key")
	if err != nil {
		return ""
	}
	return rc
}
