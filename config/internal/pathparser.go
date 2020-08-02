package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// PathParser -- parses and validates Path values
//
// we use different types of paths. we distinguish
//
// - single vs multi-value paths (.AllowMultiple)
//   If only one value is expected, the string passed to Patharser will be used as-is.
//   If multiple paths are allowed, any string starting with '[' will be parsed as JSON (parsing errors ending up in Value.Error)
// - simple paths or URLs
//   Some config paths (namely the one for ondevice.sock) allow URLs to be used.
//   PathParser will try to parse each URL and check their validity.
//
//
// When allowing URLs, we check:
// - whether the given value is a path or URL (i.e. can be parsed by url.Parse() and has a Scheme prefix)
// - simple paths will be used as-is
// - for URLs we check if their Scheme is in our Whitelist and whether they match some simple checks
//   - HTTP(S) URLs can't have a path (e.g. ondevice.sock=http://localhost:1234/foo.txt wouldn't make sense)
//   - UNIX (or file://) URLs can't have a hostname (to prevent accidents like unix://var/run/ondevice/ondevice.sock)
type PathParser struct {
	// AllowMultiple -- if true, attempt to json.Unmarshal into a string slice before validating the whole thing
	AllowMultiple bool

	// ValidSchemes -- if not nil, parse as URL and check if the Scheme matches one of the ones defined here (case insensitive -> only store lower case keys in here)
	ValidSchemes map[string]bool
}

// validatePath -- checks an individual path
func (v PathParser) validatePath(value string) error {
	if v.ValidSchemes != nil {
		var u, err = url.Parse(value)
		if err != nil {
			logrus.WithError(err).WithField("url", value).Error("PathParser: invalid URL")
			return err
		}

		if u.RawQuery != "" {
			logrus.WithField("url", value).Error("URL can't have a query string")
			return errors.New("URL query string unsupported")
		}
		if u.Fragment != "" {
			logrus.WithField("url", value).Error("URL can't have a fragment")
			return errors.New("URL fragment unsupported")
		}

		var scheme = strings.ToLower(u.Scheme)
		if !v.ValidSchemes[scheme] {
			logrus.WithField("url", value).Error("URL scheme not supported")
			return errors.New("unsupported URL scheme: " + u.Scheme)
		}

		switch scheme {
		case "http", "https":
			if u.Path != "" || u.RawPath != "" || u.Opaque != "" {
				if u.Path != "/" {
					logrus.WithField("url", value).Errorf("%s URLs should only have a hostPort, but no path", strings.ToUpper(scheme))
					return errors.New("PathParser: got path in HTTP(s) URL")
				}
			}
		case "", "file", "unix":
			if u.Host != "" {
				logrus.WithField("url", value).Error("file paths can't have a hostname portion")
				logrus.Info("  you may have used something like file://etc/motd instead of file:/etc/motd (or file:///etc/motd)")
				return errors.New("PathParser: got host in file URL")
			}
		}

		return nil
	}

	// simple path: assume it's valid
	return nil
}

// Value -- puts the parsed data into a Value
func (v PathParser) Value(raw string) ValueImpl {
	if len(raw) == 0 {
		return ValueImpl{} // empty value
	}

	if v.AllowMultiple && strings.HasPrefix(raw, "[") {
		// parse as JSON
		var rc ValueImpl

		if err := json.Unmarshal([]byte(raw), &rc.values); err != nil {
			return ValueImpl{err: fmt.Errorf("failed to parse JSON path: %s", err.Error())}
		}

		// valid JSON list -> check each individual item
		for _, value := range rc.values {
			if rc.err = v.validatePath(value); rc.err != nil {
				break
			}
		}
		return rc
	}

	// default behaviour: put it into the first slice
	return ValueImpl{
		values: []string{raw},
		err:    v.validatePath(raw),
	}
}
