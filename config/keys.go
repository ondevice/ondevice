package config

import "fmt"

// KeyClientTimeout -- specifies the timeout for HTTP requests
var KeyClientTimeout = newKey("client", "timeout", "30")

// KeyDeviceID -- represents the key where we store devId ('device.devId', defaults to '')
var KeyDeviceID = newKey("device", "dev-id", "").setRO()

// CommandRSYNC -- the path to the 'rsync' command
var CommandRSYNC = newKey("command", "rsync", "rsync")

// CommandSCP -- the path to the 'scp' command
var CommandSCP = newKey("command", "scp", "scp")

// CommandSFTP -- the path to the 'sftp' command
var CommandSFTP = newKey("command", "sftp", "sftp")

// CommandSSH -- the path to the 'ssh' command
var CommandSSH = newKey("command", "ssh", "ssh")

// PathAuthJSON -- the path to 'auth.json', relative to 'ondevice.conf'
var PathAuthJSON = newKey("path", "auth_json", "auth.json")

// PathKnownHosts -- the path to our 'known_hosts' file, relative to 'ondevice.conf'
var PathKnownHosts = newKey("path", "known_hosts", "known_hosts")

// PathOndevicePID -- the path to 'ondevice.pid', relative to 'ondevice.conf'
var PathOndevicePID = newKey("path", "ondevice_pid", "ondevice.pid")

// PathOndeviceSock -- the path to 'ondevice.sock', relative to 'ondevice.conf'
var PathOndeviceSock = newKey("path", "ondevice_sock", "ondevice.sock")

// Key -- config key struct (use the predefined Key* values when using config)
type Key struct {
	section, key, defaultValue string

	ro bool
}

// setRO -- marks configKey as being read-only (to users running 'ondevice config')
func (k Key) setRO() Key {
	k.ro = true
	allKeys[k.String()].ro = true
	return k
}

func (k Key) String() string {
	return fmt.Sprintf("%s.%s", k.section, k.key)
}

// WithDefault -- returns a modified configKey with defaultValue set to 'val'
func (k Key) WithDefault(val string) Key {
	return Key{
		section:      k.section,
		key:          k.key,
		defaultValue: val,
		ro:           k.ro,
	}
}

func newKey(section string, key string, defaultValue string) Key {
	var rc = Key{
		section:      section,
		key:          key,
		defaultValue: defaultValue,
	}
	allKeys[rc.String()] = &rc
	return rc
}

// AllKeys -- returns all defined config Keys
//
// if withReadOnly is set to false, this will filter out values not
func AllKeys(withReadOnly bool) map[string]Key {
	var rc = make(map[string]Key, len(allKeys))

	for k, v := range allKeys {
		if withReadOnly || !v.ro {
			rc[k] = *v
		}
	}

	return rc
}

// FindKey -- returns the configKey for the given string key (or nil if not found)
func FindKey(keyName string) *Key {
	return allKeys[keyName]
}

var allKeys = make(map[string]*Key)
