package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientAuth(t *testing.T) {
	setupTests()

	auth, err := GetClientAuth()
	assert.NoError(t, err)
	assert.Equal(t, "demo", auth.User())
	assert.Equal(t, "234567", auth.Key())
}

func TestClientUserAuth(t *testing.T) {
	setupTests()

	auth, err := GetClientUserAuth("demo")
	assert.NoError(t, err)
	assert.Equal(t, "demo", auth.User())
	assert.Equal(t, "234567", auth.Key())

	auth, err = GetClientUserAuth("hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello", auth.User())
	assert.Equal(t, "345678", auth.Key())

	// test case insensitivity
	auth, err = GetClientUserAuth("HeLLO")
	assert.NoError(t, err)
	assert.Equal(t, "HeLLO", auth.User())
	assert.Equal(t, "345678", auth.Key())

	// test missing user
	auth, err = GetClientUserAuth("nonexisting")
	assert.Error(t, err)
	assert.Nil(t, auth)
}

func TestDeviceAuth(t *testing.T) {
	setupTests()

	auth, err := GetDeviceAuth()
	assert.NoError(t, err)
	assert.Equal(t, "hello", auth.User())
	assert.Equal(t, "123456", auth.Key())
}
