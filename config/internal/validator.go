package internal

// Validator -- used for validating config values
type Validator interface {
	Value(raw string) Value
	Validate(value string) error
}
