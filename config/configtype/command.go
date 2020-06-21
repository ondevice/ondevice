package configtype

import "encoding/json"

// Command -- wrap settings values in this to parse them as commands
//
// if they're a valid JSON string array, it's used as value.
// otherwise, the []string slice is returned with the raw value as only element
type Command []string

// ParseCommand -- parses a Command config value (e.g. CommandSSH) and returns a string list
//
// - if t contains a valid JSON string list, it will be unmarshalled and returned
// - otherwise, returns t as the only element in the string list
//
// you'll most likely want to call Config.GetCommand() instead of calling this directly
func ParseCommand(value string) Command {
	var rc Command
	if err := json.Unmarshal([]byte(value), &rc); err != nil {
		return Command{string(value)}
	}
	return rc
}
