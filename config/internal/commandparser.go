package internal

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CommandParser -- parses and validates Command values
//
// There are two ways to specify commands
// - literal string path
// - JSON arrays (any string starting with the character '[' will be treated as such)
//   invalid JSON strings will cause validation errors
//
// In the first case, the resulting Value will simply be an array with one single item: the string itself
//
// If you want ondevice to call the command in question with predefined arguments, use the JSON variant,
// e.g: `ssh -C` (compressing data) should be declared as `["ssh", "-C"]`.
//
// to use a command starting with '[', wrap it inside a JSON array: `["["]`
type CommandParser struct{}

// Validate -- returns any errors found while parsing this Command Value
func (v CommandParser) Validate(value string) error {
	return v.Value(value).Error()
}

// Value -- parses the command config string into a config.Value
//
// if the string you pass to .Value() starts with '[', it will be parsed as JSON.
//
// Otherwise the whole string will be put inside the first (and only) array element
func (v CommandParser) Value(raw string) ValueImpl {
	if len(raw) == 0 {
		return ValueImpl{} // empty value -> use default
	}

	var rc ValueImpl
	if strings.HasPrefix(raw, "[") {
		// the value starts with '[', assume it is JSON
		if err := json.Unmarshal([]byte(raw), &rc.values); err != nil {
			rc.err = fmt.Errorf("failed to parse JSON command: %s", err.Error())
		}
	} else {
		// default behaviour: put it into the first slice
		rc.values = []string{raw}
	}
	return rc
}
