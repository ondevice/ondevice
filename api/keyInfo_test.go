package api

import (
	"testing"

	"github.com/ondevice/ondevice/config"
	"github.com/stretchr/testify/assert"
)

func TestKeyInfo(t *testing.T) {
	// check the 'demo' user's device key
	auth := NewAuth(config.NewAuth("demo", "ehb8f971h1", ""))
	info, err := GetKeyInfo(auth)
	assert.NoError(t, err)
	assert.Equal(t, "device", info.Role)
	assert.Equal(t, []string{"device"}, info.Permissions)

	// and the 'demo' user's client key
	auth = NewAuth(config.NewAuth("demo", "caxuaph5th", ""))
	info, err = GetKeyInfo(auth)
	assert.NoError(t, err)
	assert.Equal(t, "client", info.Role)
	assert.True(t, info.HasPermission("connect"))
	assert.True(t, info.HasPermission("get_properties"))
	assert.True(t, info.HasPermission("list_devices"))
	assert.False(t, info.HasPermission("device"))
	assert.Equal(t, 3, len(info.Permissions))

	if false {
		// nonexisting user
		auth = NewAuth(config.NewAuth("xxx", "blablabla", ""))
		_, err = GetKeyInfo(auth)
		assert.Error(t, err)

		// wrong auth key
		auth = NewAuth(config.NewAuth("demo", "blablabla", ""))
		_, err = GetKeyInfo(auth)
		assert.Error(t, err)
	}
}
