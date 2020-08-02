package config

// Value -- represents a parsed config value
//
// returned by config.GetValue(), and filled by the Key's parser's constraints (e.g. if the type allows multiple values, they are split up, else everything is put in [0])
//
// Has most methods act on the first element. To move to the next (i.e. take away the first), call Next()
type Value interface {
	// Len -- returns the number of (remaining) elements stored in this Value
	Len() int

	// Next -- returns a copy of this Value with the first .values entry removed
	//
	// Note that you must reassign your iterator variable to the value returned (as Value is call-by-value)
	Next() bool

	// Error -- returns type-specific validation errors
	Error() error

	// Strings -- returns all (remaining) values stored in this Value
	Strings() []string
}
