package config

// GetDeviceKey -- Returns the device's key (or an empty string if not defined)
func GetDeviceKey() string {
	rc, err := GetValue("device", "key")
	if err != nil {
		return ""
	}
	return rc
}
