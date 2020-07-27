package internal

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// PathValidator -- validates Path values
//
//
// - JSON array ()
type PathValidator struct {
	// AllowMultiple -- if true, attempt to json.Unmarshal into a string slice before validating the whole thing
	AllowMultiple bool

	// ValidSchemes -- if not nil, parse as URL and check if the Scheme matches one of the ones defined here (case insensitive -> only store lower case keys in here)
	ValidSchemes map[string]bool
}

// Validate -- checks if the given value meets our criteria
func (v PathValidator) Validate(value string) error {
	return v.Value(value).Error
}

// validatePath -- checks an individual path
func (v PathValidator) validatePath(value string) error {
	if v.ValidSchemes != nil {
		var u, err = url.Parse(value)
		if err != nil {
			logrus.WithError(err).WithField("url", value).Error("PathValidator: invalid URL")
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
					return errors.New("PathValidator: got path in HTTP(s) URL")
				}
			}
		case "", "file", "unix":
			if u.Host != "" {
				logrus.WithField("url", value).Error("file paths can't have a hostname portion")
				logrus.Info("  you may have used something like file://etc/motd instead of file:/etc/motd (or file:///etc/motd)")
				return errors.New("PathValidator: got host in file URL")
			}
		}

		return nil
	}

	// simple path: assume it's valid
	return nil
}

// Value -- puts the parsed data into a Value
func (v PathValidator) Value(raw string) Value {
	if len(raw) == 0 {
		return Value{} // empty value
	}

	if v.AllowMultiple {
		var values []string
		if err := json.Unmarshal([]byte(raw), &values); err == nil {
			// valid JSON list -> check each individual item
			for _, value := range values {
				if err = v.validatePath(value); err != nil {
					break
				}
			}
			return Value{
				values: values,
				Error:  err,
			}
		} else if strings.HasPrefix(raw, "[") {
			logrus.WithField("path", raw).WithError(err).Warn("your path looks like it should be a JSON array - but it's not formatted properly")
		}
	}

	// default behaviour: put it into the first slice
	return Value{
		values: []string{raw},
		Error:  v.validatePath(raw),
	}
}
