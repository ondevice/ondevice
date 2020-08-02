package internal

// Parser -- parses and validates config values
type Parser interface {
	Value(raw string) ValueImpl
	Validate(value string) error
}
