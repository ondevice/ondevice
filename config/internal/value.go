package internal

// Value -- represents a parsed config value
//
// returned by config.GetValue(), and filled according to each validator's constraints (e.g. if the type allows multiple values, they are split up, else everything is put in [0])
//
// Has helper methods that mostly act on the first element
//
// can be used like an iterator:
// - HasNext() returns true as long as there are values remaining
// - Next() returns a new Value with the first item removed
type Value struct {
	values []string
	err    error
}

// Error -- returns validation errors
func (v Value) Error() error { return v.err }

// HasNext -- returns true if there are more than one values remaining
func (v Value) HasNext() bool { return len(v.values) > 1 }

// Len -- returns the number of (remaining) elements stored in this Value
func (v Value) Len() int { return len(v.values) }

// Next -- returns a copy of this Value with the first .values entry removed
//
// Note that you must reassign your iterator variable to the value returned (as Value is call-by-value)
func (v Value) Next() Value {
	if len(v.values) > 0 {
		v.values = v.values[1:]
	}
	return v
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
