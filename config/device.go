package config

// GetDeviceID -- Returns the devId if available (otherwise returns an empty string)
func (c Config) GetDeviceID() string {
	var rc string
	var err error
	if rc, err = c.GetStringOld("device", "dev-id"); err != nil {
		return ""
	}
	return rc
}
