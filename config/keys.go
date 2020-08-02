package config

import (
	"fmt"

	"github.com/ondevice/ondevice/config/internal"
)

// Key -- config key struct (use the predefined Key* values when using config)
type Key struct {
	section, key, defaultValue string

	ro bool

	parser internal.Parser
}

// KeyClientTimeout -- specifies the timeout for HTTP requests
var KeyClientTimeout = regKey(Key{
	section:      "client",
	key:          "timeout",
	defaultValue: "30",
	parser:       internal.IntParser{},
})

// KeyDeviceID -- represents the key where we store devId ('device.devId', defaults to '')
var KeyDeviceID = regKey(Key{
	section: "device", key: "dev-id",
	defaultValue: "",
	ro:           true,
})

// CommandRSYNC -- the path to the 'rsync' command
var CommandRSYNC = regKey(Key{
	section: "command", key: "rsync",
	defaultValue: "rsync",
	parser:       internal.CommandParser{},
})

// CommandSCP -- the path to the 'scp' command
var CommandSCP = regKey(Key{
	section: "command", key: "scp",
	defaultValue: "scp",
	parser:       internal.CommandParser{},
})

// CommandSFTP -- the path to the 'sftp' command
var CommandSFTP = regKey(Key{
	section: "command", key: "sftp",
	defaultValue: "sftp",
	parser:       internal.CommandParser{},
})

// CommandSSH -- the path to the 'ssh' command
var CommandSSH = regKey(Key{
	section: "command", key: "ssh",
	defaultValue: "ssh",
	parser:       internal.CommandParser{},
})

// PathAuthJSON -- the path to 'auth.json', relative to 'ondevice.conf'
var PathAuthJSON = regKey(Key{
	section: "path", key: "auth_json",
	defaultValue: "auth.json",
	parser:       internal.PathParser{},
})

// PathKnownHosts -- the path to our 'known_hosts' file, relative to 'ondevice.conf'
var PathKnownHosts = regKey(Key{
	section: "path", key: "known_hosts",
	defaultValue: "known_hosts",
	parser:       internal.PathParser{},
})

// PathOndevicePID -- the path to 'ondevice.pid', relative to 'ondevice.conf'
//
// if you specify more than one, clients will try them in order. ondevice daemon will always use the first one
var PathOndevicePID = regKey(Key{
	section: "path", key: "ondevice_pid",
	defaultValue: `["ondevice.pid", "/var/run/ondevice/ondevice.pid"]`,
	parser:       internal.PathParser{AllowMultiple: true},
})

// PathOndeviceSock -- the path to 'ondevice.sock', relative to 'ondevice.conf'
//
// if you specify more than one, clients will try them in order. ondevice daemon will always use the first one
var PathOndeviceSock = regKey(Key{
	section: "path", key: "ondevice_sock",
	defaultValue: `["ondevice.sock", "unix:///var/run/ondevice/ondevice.sock"]`,
	parser: internal.PathParser{
		AllowMultiple: true,
		ValidSchemes:  map[string]bool{"": true, "file": true, "unix": true, "http": true},
	},
})

func (k Key) String() string {
	return fmt.Sprintf("%s.%s", k.section, k.key)
}

// Validate -- if the config Key has a Parser set, run it and return an error if something went wrong
func (k Key) Validate(val string) error {
	if k.parser == nil {
		return nil
	}
	return k.parser.Validate(val)
}

func regKey(key Key) Key {
	allKeys[key.String()] = &key
	return key
}

// AllKeys -- returns all defined config Keys
//
// if withReadOnly is set to false, this will filter out values that aren't writable
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
