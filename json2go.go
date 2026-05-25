package json2go

import (
	"errors"
	"regexp"
)

var (
	identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`) // TODO: support unicode

	// ErrInvalidJSON json input could not be parsed.
	ErrInvalidJSON = errors.New("invalid json")

	// ErrInvalidStructName struct name provided is not a valid Go identifier.
	ErrInvalidStructName = errors.New("invalid struct name")
)

// Transform converts a JSON string to Go struct type definitions.
//
// The structName must be a valid Go identifier (matching ^[A-Za-z_][A-Za-z0-9_]*$).
// Set includeTags to true to generate `json:"field_name"` tags on struct fields.
// Returns the Go code as a string, or an error if JSON parsing fails.
func Transform(structName, jsonStr string, includeTags bool) (string, error) {
	if !identRe.MatchString(structName) {
		return "", ErrInvalidStructName
	}

	input := unsafe.Slice(unsafe.StringData(jsonStr), len(jsonStr))
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	v, err := parser.Parse()
	if err != nil {
		return "", errors.Join(ErrInvalidJSON, err)
	}

	return NewTranspiler().Transpile(structName, v, includeTags)
}
