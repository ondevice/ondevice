package filter

import (
	"encoding/json"
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
	assert.False(t, MustMatch(dev, "empty"))
	assert.True(t, MustMatch(dev, "answer!="))

	// doesn't exist / empty (comparing them with the empty string)
	assert.False(t, MustMatch(dev, "doesntExist"))
	assert.True(t, MustMatch(dev, "doesntExist="))
	assert.False(t, MustMatch(dev, "doesntExist!="))

	assert.False(t, MustMatch(dev, "empty"))
	assert.True(t, MustMatch(dev, "empty="))
	assert.False(t, MustMatch(dev, "empty!="))

	assert.False(t, MustMatch(dev, "nullValue"))
	assert.True(t, MustMatch(dev, "nullValue="))
	assert.False(t, MustMatch(dev, "nullValue!="))

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

	// invalid filter expressions
	_, err = Matches(dev, "empty=>")
	assert.Error(t, err)
	_, err = Matches(dev, "=>foo")
	assert.Error(t, err)
}

func TestSpecial(t *testing.T) {
	var stateTs int64 = 1517852225000 // 2018-02-05T17:37:05Z
	var dev = api.Device{
		ID:        "demo.foo",
		Name:      "Some device",
		State:     "online",
		StateTs:   &stateTs,
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

	// on:id
	assert.True(t, MustMatch(dev, "on:id=demo.foo"))
	assert.True(t, MustMatch(dev, "on:id>=demo.foo"))
	assert.True(t, MustMatch(dev, "on:id>demo.fo"))
	assert.False(t, MustMatch(dev, "on:id>demo.foo"))

	// on:state
	assert.True(t, MustMatch(dev, "on:state=online"))
	assert.False(t, MustMatch(dev, "on:state=offline"))

	// on:ip
	assert.True(t, MustMatch(dev, "on:ip="))
	dev.IP = "0.1.2.3"
	assert.False(t, MustMatch(dev, "on:ip="))
	assert.True(t, MustMatch(dev, "on:ip!="))
	assert.True(t, MustMatch(dev, "on:ip=0.1.2.3"))
	assert.True(t, MustMatch(dev, "on:ip>0.1.") && MustMatch(dev, "on:ip<0.2.")) // find devices in specific IP range

	// on:createdAt:
	assert.True(t, MustMatch(dev, "on:createdAt=2018-01-15T16:55:31Z"))
	assert.True(t, MustMatch(dev, "on:createdAt>=2018")) // created this year
	assert.True(t, MustMatch(dev, "on:createdAt>2018"))  // we're still doing simple string comparison

	// on:stateTs
	assert.True(t, MustMatch(dev, "on:stateTs=2018-02-05T17:37:05Z"))
	assert.True(t, MustMatch(dev, "on:stateTs"))
	assert.True(t, MustMatch(dev, "on:stateTs!="))
	assert.False(t, MustMatch(dev, "on:stateTs<2018"))
	assert.True(t, MustMatch(dev, "on:stateTs>2018"))

	dev.StateTs = nil
	assert.False(t, MustMatch(dev, "on:stateTs"))
	assert.True(t, MustMatch(dev, "on:stateTs="))
	assert.False(t, MustMatch(dev, "on:stateTs!="))

	// other special properties
	assert.True(t, MustMatch(dev, "on:name!="))
	assert.True(t, MustMatch(dev, "on:version>=0.1"))
	assert.True(t, MustMatch(dev, "on:foo=server-defined special property"))

	// unknown special property
	assert.False(t, MustMatch(dev, "on:hello=foo"))
}

func TestTypes(t *testing.T) {
	// we get the device properties as JSON object.
	// this test makes sure the JSON deserializer works as expected (i.e. doesn't
	// unmarshal to any types we don't expect)
	var dev = api.Device{}
	assert.NoError(t, json.Unmarshal([]byte(`
		{
			"null": null,
			"bool": true,
			"smallInt": 123,
			"bigInt": 119879128371981,
			"float1": 192.0,
			"float2": 192.1,
			"intArray": [1,3,2],
			"dict": {"answer": 42}
		}
	`), &dev.Props))

	assert.True(t, MustMatch(dev, "null="))
	assert.False(t, MustMatch(dev, "null"))
	assert.True(t, MustMatch(dev, "smallInt=123"))
	assert.True(t, MustMatch(dev, "smallInt<23")) // string comparison
	assert.True(t, MustMatch(dev, "bigInt=119879128371981"))
	assert.True(t, MustMatch(dev, "bigInt<2"))
	assert.True(t, MustMatch(dev, "float1=192"))
	assert.True(t, MustMatch(dev, "float2=192.1"))

	// we can't handle arrays though
	assert.True(t, MustMatch(dev, "intArray")) // simple 'exists' expression should work

	var _, err = Matches(dev, "intArray=123")
	assert.Error(t, err)
	_, err = Matches(dev, "intArray!=")
	assert.Error(t, err)
	_, err = Matches(dev, "intArray=")
	assert.Error(t, err)

	// same goes for dicts
	assert.True(t, MustMatch(dev, "dict")) // simple 'exists' expression should work

	_, err = Matches(dev, "dict=123")
	assert.Error(t, err)
	_, err = Matches(dev, "dict!=")
	assert.Error(t, err)
	_, err = Matches(dev, "dict=")
	assert.Error(t, err)

	// TODO: object traversal
	//assert.True(t, MustMatch(dev, "dict.answer=42"))
}
