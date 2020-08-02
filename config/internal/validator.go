package internal

// Validator -- used for validating config values
type Validator interface {
	Value(raw string) ValueImpl
	Validate(value string) error
}
