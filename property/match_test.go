package property

import (
	"testing"

	"github.com/ondevice/ondevice/api"
	"github.com/stretchr/testify/assert"
)

func TestMatches(t *testing.T) {
	// TODO test path traversal (foo.bar==bak)
	// TODO test non-string values
	var dev = api.Device{
		Props: map[string]interface{}{
			"hello":     "world",
			"answer":    "42",
			"nullValue": nil, // shouldn't happen (the server's supposed to interpret 'null' as delete), but this code should treat it as the empty string
			"empty":     "",  // should be treated the same as nonexisting
		},
	}

	// exists tests (without operator)
	var _, err = Matches(dev, "")
	assert.Error(t, err)
	assert.True(t, MustMatch(dev, "hello"))
	assert.True(t, MustMatch(dev, "answer"))
	assert.False(t, MustMatch(dev, "nullValue"))
	assert.True(t, MustMatch(dev, "answer!="))

	// doesn't exist / empty (comparing them with the empty string)
	assert.True(t, MustMatch(dev, "doesntExist="))
	assert.False(t, MustMatch(dev, "doesntExist!="))
	assert.True(t, MustMatch(dev, "empty"))
	assert.False(t, MustMatch(dev, "empty!="))
	assert.True(t, MustMatch(dev, "nullValue="))

	// equality tests
	assert.True(t, MustMatch(dev, "answer==42"))
	assert.True(t, MustMatch(dev, "answer=42"))
	assert.False(t, MustMatch(dev, "answer===42")) // answer == "=42"
	assert.False(t, MustMatch(dev, "answer!=42"))
	assert.True(t, MustMatch(dev, "hello==world"))

	// comparison operators (we're doing string comparison)
	assert.False(t, MustMatch(dev, "answer<<42"))
	assert.True(t, MustMatch(dev, "answer<=42"))
	assert.True(t, MustMatch(dev, "answer<43"))
	assert.True(t, MustMatch(dev, "answer<<43"))
	assert.True(t, MustMatch(dev, "answer<5"))
	assert.False(t, MustMatch(dev, "answer< 43")) // is compared to " 43" (with starting space char)
	assert.True(t, MustMatch(dev, "hello>World")) // values are compared case sensitively

	assert.True(t, MustMatch(dev, "answer>>345"))
	assert.True(t, MustMatch(dev, "answer>=345"))
	assert.True(t, MustMatch(dev, "answer>41"))
	assert.False(t, MustMatch(dev, "answer>42"))
	assert.False(t, MustMatch(dev, "answer>43"))

	// unsupported operator
	_, err = Matches(dev, "empty=>")
	assert.Error(t, err)
}

func TestSpecial(t *testing.T) {
	var dev = api.Device{
		ID:        "demo.foo",
		Name:      "Some device",
		State:     "online",
		Version:   "ondevice v0.1.2",
		CreatedAt: 1516035331000, // 2018-01-15T16:55:31Z
		Props: map[string]interface{}{
			"hello":     "world",
			"answer":    "42",
			"nullValue": nil, // shouldn't happen (the server's supposed to interpret 'null' as delete), but this code should treat it as the empty string
			"empty":     "",  // should be treated the same as nonexisting
			"on:foo":    "server-defined special property",
		},
	}

	// TODO test matching unqualified IDs
	assert.True(t, MustMatch(dev, "on:id=demo.foo"))
	assert.True(t, MustMatch(dev, "on:id>=demo.foo"))
	assert.True(t, MustMatch(dev, "on:id>demo.fo"))
	assert.False(t, MustMatch(dev, "on:id>demo.foo"))

	// state
	assert.True(t, MustMatch(dev, "on:state=online"))
	assert.False(t, MustMatch(dev, "on:state=offline"))

	// IP
	assert.True(t, MustMatch(dev, "on:ip="))
	dev.IP = "0.1.2.3"
	assert.False(t, MustMatch(dev, "on:ip="))
	assert.True(t, MustMatch(dev, "on:ip!="))
	assert.True(t, MustMatch(dev, "on:ip=0.1.2.3"))
	assert.True(t, MustMatch(dev, "on:ip>0.1.") && MustMatch(dev, "on:ip<0.2.")) // find devices in specific IP range

	// CreatedAt:
	assert.True(t, MustMatch(dev, "on:createdAt=2018-01-15T16:55:31Z"))
	assert.True(t, MustMatch(dev, "on:createdAt>=2018")) // created this year
	assert.True(t, MustMatch(dev, "on:createdAt>2018"))  // we're still doing simple string comparison

	// other special properties
	assert.True(t, MustMatch(dev, "on:name!="))
	assert.True(t, MustMatch(dev, "on:version>=0.1"))
	assert.True(t, MustMatch(dev, "on:foo=server-defined special property"))

	// unknown special property
	_, err := Matches(dev, "on:hello=foo")
	assert.Error(t, err)
}
