package internal

// Validator -- used for validating config values
type Validator interface {
	Validate(value string) error
}