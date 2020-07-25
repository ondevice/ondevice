package internal

// Value -- represents a parsed config value
//
// returned by Validator.Value(), and filled according to the validator's constraints (e.g. if the type allows multiple values, they are split up, else everything is put in [0])
//
// has helper methods that mostly act on the first element
//
// can be used like an iterator:
// - HasNext() returns true as long as there are values remaining
// - Next() returns a new Value with the first item removed
//
// immutable if you only use the helper methods
type Value struct {
	values []string
	Error  error
}

// returns true if there are
func (v Value) HasNext() bool { return len(v.values) > 1 }

func (v Value) Next() Value {
	if len(v.values) > 0 {
		v.values = v.values[1:]
	}
	return v
}

func (v Value) String(defaultValue string) string {
	if len(v.values) == 0 {
		return defaultValue
	}
	return v.values[0]
}
