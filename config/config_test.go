package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTests() {
	_configPath = "../testData/ondevice.conf"
}

func TestPathOverride(t *testing.T) {
	_configPath = "/tmp/notfound/ondevice.conf"
	var cfg, err = Load()
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
	assert.Equal(t, _configPath, cfg.path)

	// even though ondevice.conf wasn't found, we expect 'auth.json' to be in the same directory
	assert.Equal(t, "/tmp/notfound/auth.json", cfg.GetFilePath(PathAuthJSON), "config path override failed")

	// update path to auth.json
	cfg.SetValue(PathAuthJSON, "/etc/ondevice/auth.json")
	assert.Equal(t, "/etc/ondevice/auth.json", cfg.GetFilePath(PathAuthJSON), "config path override failed")

	// make sure MustLoad() doesn't fail on FileNotExists
	MustLoad()
}

func TestGetString(t *testing.T) {
	setupTests()

	var cfg, err = Load()
	assert.NoError(t, err)

	var val = cfg.GetString(KeyClientTimeout)
	assert.Equal(t, "123", val)

	// test case insensitivity
	val = cfg.GetString(Key{section: "CliEnt", key: "tIMeout"})
	assert.Equal(t, "123", val)

	// test missing section
	val = cfg.GetString(Key{section: "client_", key: "timeout", defaultValue: "notFound"})
	assert.Equal(t, "notFound", val)

	// test missing key
	val = cfg.GetString(Key{section: "client", key: "timeout_", defaultValue: "notFound"})
	assert.Equal(t, "notFound", val)

	// test missing config file
	_configPath = "/tmp/nonexisting/ondevice.conf"
	cfg, err = Load()
	assert.Error(t, err)
	val = cfg.GetString(KeyClientTimeout)
	assert.Error(t, err)
	assert.Equal(t, "30", val)
}

func TestGetInt(t *testing.T) {
	setupTests()

	var cfg, err = Load()
	assert.NoError(t, err)

	var val = cfg.GetInt(KeyClientTimeout)
	assert.Equal(t, 123, val)

	// test case insensitivity
	val = cfg.GetInt(Key{section: "CliEnt", key: "tIMeout"})
	assert.Equal(t, 123, val)

	// test missing section
	val = cfg.GetInt(Key{section: "client_", key: "timeout", defaultValue: "42"})
	assert.Equal(t, 42, val)

	// test missing key
	val = cfg.GetInt(Key{section: "client", key: "timeout_", defaultValue: "42"})
	assert.Equal(t, 42, val)

	// test invalid integer value
	var key = Key{section: "client", key: "invalidTimeout", defaultValue: "42"}
	val = cfg.GetInt(key)
	assert.Equal(t, 42, val)
	assert.Equal(t, "7b", cfg.GetString(key))

	// so apparently strconf.ParseInt() supports hex/oct/binary stuff as well...
	cfg.SetValue(KeyClientTimeout, "0x7a") // hex for 122
	assert.Equal(t, 122, cfg.GetInt(KeyClientTimeout))

	cfg.SetValue(KeyClientTimeout, "0o171") // oct for 121
	assert.Equal(t, 121, cfg.GetInt(KeyClientTimeout))

	cfg.SetValue(KeyClientTimeout, "0b1111000") // oct for 120
	assert.Equal(t, 120, cfg.GetInt(KeyClientTimeout))

	// test missing config file
	_configPath = "/tmp/nonexisting/ondevice.conf"
	cfg, err = Load()
	assert.Error(t, err)
	val = cfg.GetInt(KeyClientTimeout)
	assert.Error(t, err)
	assert.Equal(t, 30, val)

}
