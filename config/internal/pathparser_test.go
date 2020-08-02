package internal

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseSinglePathWithoutURLs -- default parser without URL support (e.g.: config.PathAuthJSON)
func TestParseSinglePathWithoutURLs(t *testing.T) {
	var assert = assert.New(t)
	var parser = PathParser{}

	var value = parser.Value("")
	assert.NoError(value.Error())
	assert.Empty(value.values)
	assert.Equal("--defaultValue--", value.String("--defaultValue--"))
	assert.False(value.Next())

	// JSON won't be parsed
	var str = "[\"auth.json\", \"file:/etc/motd\"]"
	value = parser.Value(str)
	assert.NoError(value.Error())
	assert.Len(value.values, 1)
	assert.Equal(str, value.String("--defaultValue--"))
	assert.False(value.Next())

	// URLs will be treated as simple path names - resulting in file system error messages downstream
	str = "https://ondevice.io/index.html"
	value = parser.Value(str)
	assert.NoError(value.Error())
	assert.Len(value.values, 1)
	assert.Equal(str, value.String("--defaultValue--"))
	assert.False(value.Next())
}

// TestParseSinglePathWithURLs -- single paths or URLs
func TestParseSinglePathWithURLs(t *testing.T) {
	var assert = assert.New(t)
	var parser = PathParser{
		ValidSchemes: map[string]bool{"": true, "unix": true, "http": true},
	}

	var value = parser.Value("")
	assert.NoError(value.Error())
	assert.Empty(value.values)
	assert.Equal("--defaultValue--", value.String("--defaultValue--"))
	assert.False(value.Next())

	// single valid URL
	var str = "http://localhost:8080/"
	value = parser.Value(str)
	assert.NoError(value.Error())
	assert.Len(value.values, 1)
	assert.Equal(str, value.String("--defaultValue--"))
	assert.False(value.Next())

	// single URL with invalid schema (.Value will still return the stored value even if it's technically invalid)
	str = "https://ondevice.io/index.html"
	value = parser.Value(str)
	assert.Error(value.Error())
	assert.Len(value.values, 1)
	assert.Equal(str, value.String("--defaultValue--"))
	assert.False(value.Next())

	// JSON values will cause URL validation errors
	str = "[\"https://ondevice.io/index.html\"]"
	value = parser.Value(str)
	assert.Error(value.Error())
	assert.Len(value.values, 1)
	assert.Equal(str, value.String("--defaultValue--"))
	assert.False(value.Next())

	// for http sockets, the path must be empty
	str = "http://localhost/index.html"
	value = parser.Value(str)
	assert.Error(value.Error())
	assert.Len(value.values, 1)
	assert.Equal(str, value.String("--defaultValue--"))
	assert.False(value.Next())

	// for unix sockets, the hostname must be empty
	str = "unix://var/run/foo.sock"
	value = parser.Value(str)
	assert.Error(value.Error())
	assert.Len(value.values, 1)
	assert.Equal(str, value.String("--defaultValue--"))
}

// TestValidateMultiPathWithoutURLs -- multiple file paths (no URL parsing)
func TestValidateMultiPathWithoutURLs(t *testing.T) {
	var assert = assert.New(t)
	var parser = PathParser{
		AllowMultiple: true,
	}

	var value = parser.Value("")
	assert.NoError(value.Error())
	assert.Empty(value.values)
	assert.Equal("--defaultValue--", value.String("--defaultValue--"))
	assert.False(value.Next())

	// single valid path
	var str = "/var/run/ondevice.pid"
	value = parser.Value(str)
	assert.NoError(value.Error())
	assert.Len(value.values, 1)
	assert.Equal(str, value.String("--defaultValue--"))
	assert.False(value.Next())

	// single URL -> no error but won't be able to read/write the files
	str = "http://localhost:8080/"
	value = parser.Value(str)
	assert.NoError(value.Error())
	assert.Len(value.values, 1)
	assert.Equal(str, value.String("--defaultValue--"))
	assert.False(value.Next())

	// valid JSON string array (since we don't do URL parsing, this won't cause errors)
	str = "[\"~/.config/ondevice/ondevice.sock\", \"/var/run/ondevice.sock\", \"http://localhost:8080\"]"
	value = parser.Value(str)
	assert.NoError(value.Error())
	assert.Len(value.values, 3)
	assert.Equal("~/.config/ondevice/ondevice.sock", value.String("--defaultValue--"))

	assert.True(value.Next())
	assert.Equal("/var/run/ondevice.sock", value.String("--defaultValue--"))

	assert.True(value.Next())
	assert.Equal("http://localhost:8080", value.String("--defaultValue--"))
	assert.False(value.Next())

	// valid single-item JSON array
	str = "[\"https://ondevice.io/index.html\"]"
	value = parser.Value(str)
	assert.NoError(value.Error())
	assert.Len(value.values, 1)
	assert.Equal("https://ondevice.io/index.html", value.String("--defaultValue--"))
	assert.False(value.Next())

	// invalid json -> error
	str = "[\"https://ondevice.io/index.html\""
	value = parser.Value(str)
	assert.Error(value.Error())
	assert.Empty(value.values)
	assert.Equal("--defaultValue--", value.String("--defaultValue--"))

	// valid JSON, but not a list -> should be interpreted as-is
	str = "\"https://ondevice.io/index.html\""
	value = parser.Value(str)
	assert.NoError(value.Error())
	assert.Len(value.values, 1)
	assert.Equal(str, value.String("--defaultValue--"))
	assert.False(value.Next())

	// valid JSON list, but not of strings -> error
	str = "[\"https://ondevice.io/index.html\", 1,2,3]"
	value = parser.Value(str)
	assert.Error(value.Error())
	assert.Empty(value.values)
	assert.Equal("--defaultValue--", value.String("--defaultValue--"))
}

// TestParseURL -- tests some edge cases of URL parsing with url.Parse()
func TestParseURL(t *testing.T) {
	var assert = assert.New(t)
	var u, err = url.Parse("")
	assert.NoError(err)
	assert.Equal(url.URL{}, *u)

	u, err = url.Parse("/etc/motd")
	assert.NoError(err)
	assert.Equal(url.URL{Path: "/etc/motd"}, *u)

	u, err = url.Parse("ondevice.conf")
	assert.NoError(err)
	assert.Equal(url.URL{Path: "ondevice.conf"}, *u)

	u, err = url.Parse("~/.config/ondevice/ondevice.conf")
	assert.NoError(err)
	assert.Equal(url.URL{Path: "~/.config/ondevice/ondevice.conf"}, *u)

	// file:// schema
	u, err = url.Parse("file:///home/user/.config/ondevice/ondevice.conf")
	assert.NoError(err)
	assert.Equal(url.URL{Scheme: "file", Path: "/home/user/.config/ondevice/ondevice.conf"}, *u)

	u, err = url.Parse("file:~/.config/ondevice/ondevice.conf")
	assert.NoError(err)
	assert.Equal(url.URL{Scheme: "file", Opaque: "~/.config/ondevice/ondevice.conf"}, *u)

	u, err = url.Parse("file:/home/user/.config/ondevice/ondevice.conf")
	assert.NoError(err)
	assert.Equal(url.URL{Scheme: "file", Path: "/home/user/.config/ondevice/ondevice.conf"}, *u)

	// make sure query params and fragment (? and #) are stored in the right fields
	u, err = url.Parse("file:/etc/motd?test=123#fragmeNt")
	assert.Equal(url.URL{Scheme: "file", Path: "/etc/motd", RawQuery: "test=123", Fragment: "fragmeNt"}, *u)

	// problem: Host may not be set when Scheme is file or unix
	u, err = url.Parse("file://home/user/.config/ondevice/ondevice.conf")
	assert.NoError(err)
	assert.Equal(url.URL{Scheme: "file", Host: "home", Path: "/user/.config/ondevice/ondevice.conf"}, *u)

	// these are functionally identical
	u1, e1 := url.Parse("file:///etc/motd")
	u2, e2 := url.Parse("file:/etc/motd")
	assert.NoError(e1)
	assert.NoError(e2)
	assert.Equal(u1, u2)

	// minor problem: if we pass raw JSON data to url.Parse() it won't cause an error (it'll simply end up in .Path and .RawPath)
	// (we always try to unmarshall first, so this shouldn't be too bad)
	var str = "\"/etc/motd\""
	u, err = url.Parse(str)
	assert.NoError(err)
	assert.Equal(u.Path, u.RawPath)
	assert.Equal(str, u.Path)

	str = "[\"~/.config/ondevice.sock\", \"file:/var/run/ondevice/ondevice.sock\"]"
	u, err = url.Parse(str)
	assert.NoError(err)
	assert.Equal(u.Path, u.RawPath)
	assert.Equal(str, u.Path)

	// what we've learned
	//
	// - use UNIX sockets if schema is '', 'file' or 'unix'
	// - use HTTP socket if schema is 'http' (and maybe 'tcp'?)
	// - we won't allow HTTPS for now, but that might change
	//
	// - for UNIX sockets:
	//   - if .Scheme is empty, use the RAW string instead of the parsed URL
	//   - make sure that .Host is empty (to avoid the file://path/to/file issue being parsed as .Host='path')
	//   - the path may be in .Path or .Opaque
	// - for HTTP sockets make sure .Path and .Opaque are empty
	// - warn if .Query or .Fragment aren't empty
}

// TODO test with windows paths
/*func TestFileURLsOnWindows(t *testing.T) {
		u, err = url.Parse("c:/Documents and Settings/hello.txt")
		assert.NoError(err)
		assert.Equal(url.URL{Scheme: "c", Path: "/Documents and Settings/hello.txt", RawPath: "/Documents and Settings/hello.txt"}, *u)

}*/
