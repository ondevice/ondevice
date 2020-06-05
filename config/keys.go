package config

import "fmt"

// KeyClientTimeout -- specifies the timeout for HTTP requests
var KeyClientTimeout = newKey("client", "timeout", "0")

// KeyDeviceID -- represents the key where we store devId ('device.devId', defaults to '')
var KeyDeviceID = newKey("device", "devId", "").setRO()

// KeySSHCommand -- the path to the 'ssh' command
var KeySSHCommand = newKey("command", "ssh", "ssh")

type configKey struct {
	section, key, defaultValue string

	ro bool
}

// setRO -- marks configKey as being read-only (to users running 'ondevice config')
func (k *configKey) setRO() *configKey {
	k.ro = true
	return k
}

func (k configKey) String() string {
	return fmt.Sprintf("%s.%s", k.section, k.key)
}

// returns a modified configKey with defaultValue set to 'val'
func (k configKey) WithDefault(val string) configKey {
	return configKey{
		section:      k.section,
		key:          k.key,
		defaultValue: val,
		ro:           k.ro,
	}
}

func newKey(section string, key string, defaultValue string) *configKey {
	var rc = configKey{
		section:      section,
		key:          key,
		defaultValue: defaultValue,
	}
	allKeys[rc.String()] = &rc
	return &rc
}

var allKeys = make(map[string]*configKey)
