package config

// GetDeviceID -- Returns the devId if available (otherwise returns an empty string)
func GetDeviceID() string {
	var cfg, err = Read()
	if err != nil {
		return ""
	}

	var rc string
	if rc, err = cfg.GetString("device", "dev-id"); err != nil {
		return ""
	}
	return rc
}

// GetDeviceKey -- Returns the device's key (or an empty string if not defined)
func GetDeviceKey() string {
	var cfg, err = Read()
	if err != nil {
		return ""
	}

	var rc string
	if rc, err = cfg.GetString("device", "key"); err != nil {
		return ""
	}
	return rc
}
