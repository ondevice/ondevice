package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTests() {
	_configPath = "../testData/ondevice.conf"
}

func TestPathOverride(t *testing.T) {
	_configPath = "/tmp/ondevice_test/ondevice.conf"
	assert.Equal(t, "/tmp/ondevice_test/test.txt", GetConfigPath("test.txt"), "Config path override failed")
}

func TestGetString(t *testing.T) {
	setupTests()

	var cfg, err = Read()
	assert.NoError(t, err)

	user, err := cfg.GetString("device", "user")
	assert.NoError(t, err)
	assert.Equal(t, "hello", user)

	// test case insensitivity
	user, err = cfg.GetString("devIce", "User")
	assert.NoError(t, err)
	assert.Equal(t, "hello", user)

	// test missing section
	user, err = cfg.GetString("device_", "user")
	assert.Error(t, err)
	assert.Equal(t, "", user)

	// test missing key
	user, err = cfg.GetString("device", "user_")
	assert.Error(t, err)
	assert.Equal(t, "", user)

	// test missing config file
	_configPath = "/tmp/nonexisting/ondevice.conf"
	cfg, err = Read()
	assert.Error(t, err)
	user, err = cfg.GetString("device", "user_")
	assert.Error(t, err)
	assert.Equal(t, "", user)
}
