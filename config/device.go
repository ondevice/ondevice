package config

// GetDeviceID -- Returns the devId if available (otherwise returns an empty string)
func (c Config) GetDeviceID() string {
	var rc string
	var err error
	if rc, err = c.GetString("device", "dev-id"); err != nil {
		return ""
	}
	return rc
}

// GetDeviceKey -- Returns the device's key (or an empty string if not defined)
func (c Config) GetDeviceKey() string {
	var rc string
	var err error
	if rc, err = c.GetString("device", "key"); err != nil {
		return ""
	}
	return rc
}
