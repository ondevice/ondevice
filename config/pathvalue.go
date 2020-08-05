package config

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/ondevice/ondevice/config/internal"
	"github.com/sirupsen/logrus"
)

// PathValue -- Value for file paths and URLs (call Config.GetPath())
type PathValue struct {
	internal.ValueImpl

	// relative file paths are
	configPath string
	AllowURLs  bool
	// NoExpandTilde bool // maybe we want this to be configurable too?
}

// GetPath -- returns the path part of the given config value
//
// - if the value is schema-less (or .AllowURLS is false), returns it as-is
// - for file/unix sockets, returns url.URL.Path or .Opaque
// - for other sockets returns the empty string
//
// if parsing the URL failed, also returns the empty string
func (v PathValue) GetPath() string {
	var rc = v.String("")
	if !v.AllowURLs {
		return rc
	}

	var u, err = url.Parse(rc)
	if err != nil {
		logrus.WithError(err).WithField("url", rc).Error("failed to parse PathValue URL")
		return ""
	}

	switch u.Scheme {
	case "":
		return rc
	case "file", "unix":
		if u.Opaque != "" {
			return u.Opaque
		}
		return u.Path
	default:
		return ""
	}
}

// GetAbsolutePath -- returns an absolute version of path
//
// wraps GetPath() and
// - returns absolute paths (as well as the empty string) as-is
// - expands a leading '~' to the user's HOME directory
// - returns relative paths as absolute (relative to .configPath)
func (v PathValue) GetAbsolutePath() string {
	return v.makeAbsolute(v.GetPath())
}

// GetAbsoluteURL -- parses the current value as URL and makes it absolute if necessary
func (v PathValue) GetAbsoluteURL() *url.URL {
	var str = v.String("")
	if str == "" {
		return nil
	}

	var u, err = url.Parse(str)
	if err != nil {
		logrus.WithError(err).Error("PathValue.GetAbsoluteURL(): failed to parse URL")
		return nil
	}

	switch u.Scheme {
	case "http", "https":
		return u
	default:
		u.Opaque = v.makeAbsolute(u.Opaque)
		u.Path = v.makeAbsolute(u.Path)
	}

	return u
}

// makeAbsolute -- makes the given path absolute
//
// either by expanding a leading '~' or by making it relative to .configPath
func (v PathValue) makeAbsolute(path string) string {
	if path == "" {
		return ""
	}

	if strings.HasPrefix(path, "~") {
		var err error
		if path, err = homedir.Expand(path); err != nil {
			logrus.WithError(err).Error("PathValue: failed to expand homedir")
			return ""
		}
	} else if !filepath.IsAbs(path) {
		// go's filepath.Join ALWAYS concats the two, no matter if the second one is absolute or relative
		var dir = filepath.Dir(v.configPath)
		path = filepath.Join(dir, path)
	}

	return path
}
