package json2go

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func field(name, type_ string, json_ ...string) string {
	if strings.Contains(type_, "struct") {
		return fmt.Sprintf("\n%s %s", name, type_)
	}

	tag := strings.ToLower(name)
	if len(json_) == 1 {
		tag = json_[0]
	}
	return fmt.Sprintf("\n%s %s `json:\"%s\"`", name, type_, tag)
}

func TestTransformer_Transform(t *testing.T) {
	tests := map[string]struct {
		input  string
		output string
		err    error
	}{
		"simple object": {
			input: `{"name": "Olex", "active": true, "age": 420}`,
			output: "type Out struct {" +
				field("Active", "bool") +
				field("Age", "int") +
				field("Name", "string") +
				"\n}",
		},
		"invalid json": {
			err:   ErrInvalidJSON,
			input: `{"invalid":json}`,
		},
		"snake_case to CamelCase": {
			input: `{"first_name": "Bob", "last_name": "Bobberson"}`,
			output: "type Out struct {" +
				field("FirstName", "string", "first_name") +
				field("LastName", "string", "last_name") +
				"\n}",
		},
		"nested object and array": {
			input: `{"user": {"name": "Alice", "score": 95.5}, "tags": ["go", "json"]}`,
			output: "type Out struct {" +
				field("Tags", "[]string") +
				field("User", "struct {") +
				field("Name", "string") +
				field("Score", "float64") +
				"\n} `json:\"user\"`" +
				"\n}",
		},
		"empty nested object": {
			input: `{"user": {}}`,
			output: "type Out struct {" +
				field("User", "struct {") +
				"\n} `json:\"user\"`" +
				"\n}",
		},
		"array of object": {
			input: `[{"name": "John"}, {"name": "Jane"}]`,
			output: "type Out []struct {" +
				field("Name", "string") +
				"\n}",
		},
		"empty array": {
			input: `{"items": []}`,
			output: "type Out struct {" +
				field("Items", "[]any") +
				"\n}",
		},
		"null": {
			input: `{"item": null}`,
			output: `type Out struct {` +
				field("Item", "any") +
				"\n}",
		},
		"numbers": {
			input: `{"pos": 123, "neg": -321, "float": 420.69}`,
			output: "type Out struct {" +
				field("Float", "float64") +
				field("Neg", "int") +
				field("Pos", "int") +
				"\n}",
		},
	}

	trans := NewTransformer()
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			result, err := trans.Transform("Out", tt.input)
			assertEqualErr(t, tt.err, err)
			assertEqual(t, tt.output, result)
		})
	}
}

func assertEqualErr(t *testing.T, expected, actual error) {
	t.Helper()
	if (expected != nil || actual != nil) && errors.Is(expected, actual) {
		t.Errorf("expected: %v, got: %v\n", expected, actual)
	}
}

func assertEqual[T any](t *testing.T, expected, actual T) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected: %v, got: %v\n", expected, actual)
	}
}
