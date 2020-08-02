package internal

// Value -- represents a parsed config value
//
// returned by config.GetValue(), and parsed depending on the key type
// (e.g. if the type allows multiple values, they are split up, else everything is put in [0])
//
// Most methods act on the first element
//
// can be used like an iterator using Next()
type Value struct {
	values []string
	err    error
}

// Error -- returns validation errors
func (v Value) Error() error { return v.err }

// Len -- returns the number of (remaining) elements stored in this Value
func (v Value) Len() int { return len(v.values) }

// Next -- moves the iterator to the next value
//
// returns false if there are no more values left
func (v *Value) Next() bool {
	if len(v.values) > 0 {
		v.values = v.values[1:]
	}
	return len(v.values) > 0
}

// String -- returns the first item stored in this Value - or defaultValue if empty
func (v Value) String(defaultValue string) string {
	if len(v.values) == 0 {
		return defaultValue
	}
	return v.values[0]
}

// Strings -- returns all (remaining) values stored in this Value
func (v Value) Strings() []string {
	return v.values
}
