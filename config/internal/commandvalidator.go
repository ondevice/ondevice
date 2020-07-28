package internal

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
)

// CommandValidator -- constructs a config.Value from a string
type CommandValidator struct{}

// Validate -- always returns nil
func (v CommandValidator) Validate(value string) error {
	return v.Value(value).Error
}

// Value -- parses the command config string into a config.Value
//
// if the string you pass to .Value() is a valid JSON string array, it'll be parsed and split into its components.
// otherwise it'll
func (v CommandValidator) Value(raw string) Value {
	var args []string

	if len(raw) == 0 {
		return Value{} // empty value -> use default
	}

	if err := json.Unmarshal([]byte(raw), &args); err == nil {
		// valid JSON list - use as is
		return Value{
			values: args,
		}
	} else if strings.HasPrefix(raw, "[") {
		logrus.WithField("cmd", raw).WithError(err).Error("your command looks like it's in JSON format but contains errors")
		return Value{Error: err}
	}

	// default behaviour: put it into the first slice
	return Value{
		values: []string{raw},
	}
}