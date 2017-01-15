package config

import "github.com/stretchr/testify/assert"
import "testing"

func TestClientAuth(t *testing.T) {
	setupTests()

	user, auth, err := GetClientAuth()
	assert.Equal(t, "demo", user)
	assert.Equal(t, "234567", auth)
	assert.Nil(t, err)
}

func TestClientUserAuth(t *testing.T) {
	setupTests()

	user, auth, err := GetClientUserAuth("demo")
	assert.Equal(t, "demo", user)
	assert.Equal(t, "234567", auth)
	assert.Nil(t, err)

	user, auth, err = GetClientUserAuth("hello")
	assert.Equal(t, "hello", user)
	assert.Equal(t, "345678", auth)
	assert.Nil(t, err)

	// test case insensitivity
	user, auth, err = GetClientUserAuth("HeLLO")
	assert.Equal(t, "HeLLO", user)
	assert.Equal(t, "345678", auth)
	assert.Nil(t, err)

	// test missing user
	user, auth, err = GetClientUserAuth("nonexisting")
	assert.Equal(t, "", user)
	assert.Equal(t, "", auth)
	assert.NotNil(t, err)
}

func TestDeviceAuth(t *testing.T) {
	setupTests()

	user, auth, err := GetDeviceAuth()
	assert.Equal(t, "hello", user)
	assert.Equal(t, "123456", auth)
	assert.Nil(t, err)
}
