package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTests() {
	_configPath = "../testData"
}

func TestPathOverride(t *testing.T) {
	_configPath = "/tmp/ondevice_test/"
	assert.Equal(t, "/tmp/ondevice_test/test.txt", GetConfigPath("test.txt"), "Config path override failed")
}

func TestGetValue(t *testing.T) {
	setupTests()

	user, err := GetValue("device", "user")
	assert.Equal(t, "hello", user)
	assert.Nil(t, err)

	// test case insensitivity
	user, err = GetValue("devIce", "User")
	assert.Equal(t, "hello", user)
	assert.Nil(t, err)

	// test missing section
	user, err = GetValue("device_", "user")
	assert.Equal(t, "", user)
	assert.NotNil(t, err)

	// test missing key
	user, err = GetValue("device", "user_")
	assert.Equal(t, "", user)
	assert.NotNil(t, err)

	// test missing config file
	_configPath = "/tmp/nonexisting/"
	user, err = GetValue("device", "user_")
	assert.Equal(t, "", user)
	assert.NotNil(t, err)
}
