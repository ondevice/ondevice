package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientAuth(t *testing.T) {
	setupTests()

	auth, err := LoadAuth().GetClientAuth()
	assert.NoError(t, err)
	assert.Equal(t, "demo", auth.User())
	assert.Equal(t, "234567", auth.Key())
}

func TestClientUserAuth(t *testing.T) {
	setupTests()

	var a = LoadAuth()

	auth, err := a.GetClientAuthForUser("demo")
	assert.NoError(t, err)
	assert.Equal(t, "demo", auth.User())
	assert.Equal(t, "234567", auth.Key())

	auth, err = a.GetClientAuthForUser("hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello", auth.User())
	assert.Equal(t, "345678", auth.Key())

	// test case insensitivity
	auth, err = a.GetClientAuthForUser("HeLLO")
	assert.NoError(t, err)
	assert.Equal(t, "hello", auth.User())
	assert.Equal(t, "345678", auth.Key())

	// test missing user
	auth, err = a.GetClientAuthForUser("nonexisting")
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
