package rest

import "testing"
import "github.com/stretchr/testify/assert"

func TestKeyInfo(t *testing.T) {
	auth := CreateAuth("demo", "ehb8f971h1")
	info, err := GetKeyInfo(auth)
	assert.Nil(t, err)
	assert.Equal(t, []string{"device"}, info)

	// nonexisting user
	auth = CreateAuth("xxx", "blablabla")
	_, err = GetKeyInfo(auth)
	assert.NotNil(t, err)

	// wrong auth key
	auth = CreateAuth("demo", "blablabla")
	_, err = GetKeyInfo(auth)
	assert.NotNil(t, err)
}
