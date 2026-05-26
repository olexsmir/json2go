package json2go

import (
	"errors"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

var (
	// ErrInvalidJSON json input could not be parsed.
	ErrInvalidJSON = errors.New("invalid json")

	// ErrInvalidStructName struct name provided is not a valid Go identifier.
	ErrInvalidStructName = errors.New("invalid struct name")
)

// Transform converts a JSON string to Go struct type definitions.
//
// The structName must be a valid Go identifier.
// Set includeTags to true to generate `json:"field_name"` tags on struct fields.
// Returns the Go code as a string, or an error if JSON parsing fails.
func Transform(structName, jsonStr string, includeTags bool) (string, error) {
	if !isValidIdentifier(structName) {
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

func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	r, size := utf8.DecodeRuneInString(s)
	if !unicode.IsLetter(r) && r != '_' {
		return false
	}

	for _, r := range s[size:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}

	return true
}
