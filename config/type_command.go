package config

import "encoding/json"

// CommandValue -- wrap settings values in this to parse them as commands
//
// if they're a valid JSON string array, it's used as value.
// otherwise, the []string slice is returned with the raw value as only element
type CommandValue string

// Value -- parses a command config value (e.g. CommandSSH) and returns a string list
//
// - if t contains a valid JSON string list, it will be unmarshalled and returned
// - otherwise, returns t as the only element in the string list
func (v CommandValue) Value() []string {
	var rc []string
	if err := json.Unmarshal([]byte(v), &rc); err != nil {
		return []string{string(v)}
	}
	return rc
}
