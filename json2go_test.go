package json2go

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func field(indentLvl int, name, type_ string, json_ ...string) string {
	indent := strings.Repeat("\t", indentLvl)
	if strings.Contains(type_, "struct") {
		return fmt.Sprintf("\n%s%s %s", indent, name, type_)
	}

	tag := strings.ToLower(name)
	if len(json_) == 1 {
		tag = json_[0]
	}
	return fmt.Sprintf("\n%s%s %s `json:\"%s\"`", indent, name, type_, tag)
}

func TestTransformer_Transform(t *testing.T) {
	tests := map[string]struct {
		input      string
		output     string
		structName string
		err        error
	}{
		"simple object": {
			input: `{"name": "Olex", "active": true, "age": 420}`,
			output: "type Out struct {" +
				field(1, "Active", "bool") +
				field(1, "Age", "int") +
				field(1, "Name", "string") +
				"\n}",
		},
		"invalid json": {
			err:   ErrInvalidJSON,
			input: `{"invalid":json}`,
		},
		"invalid struct name, starts with number": {
			err:        ErrInvalidStructName,
			structName: "1Name",
		},
		"invalid struct name, has space": {
			err:        ErrInvalidStructName,
			structName: "Name Name2",
		},
		"invalid struct name, has non letter/number": {
			err:        ErrInvalidStructName,
			structName: "Name$",
		},
		"snake_case to CamelCase": {
			input: `{"first_name": "Bob", "last_name": "Bobberson"}`,
			output: "type Out struct {" +
				field(1, "FirstName", "string", "first_name") +
				field(1, "LastName", "string", "last_name") +
				"\n}",
		},
		"nested object and array": {
			input: `{"user": {"name": "Alice", "score": 95.5}, "tags": ["go", "json"]}`,
			output: "type Out struct {" +
				field(1, "Tags", "[]string") +
				field(1, "User", "struct {") +
				field(2, "Name", "string") +
				field(2, "Score", "float64") +
				"\n\t} `json:\"user\"`" +
				"\n}",
		},
		"empty nested object": {
			input: `{"user": {}}`,
			output: "type Out struct {" +
				field(1, "User", "struct {") +
				"\n\t} `json:\"user\"`" +
				"\n}",
		},
		"array of object": {
			input: `[{"name": "John"}, {"name": "Jane"}]`,
			output: "type Out []struct {" +
				field(1, "Name", "string") +
				"\n}",
		},
		"empty array": {
			input: `{"items": []}`,
			output: "type Out struct {" +
				field(1, "Items", "[]any") +
				"\n}",
		},
		"null": {
			input: `{"item": null}`,
			output: `type Out struct {` +
				field(1, "Item", "any") +
				"\n}",
		},
		"numbers": {
			input: `{"pos": 123, "neg": -321, "float": 420.69}`,
			output: "type Out struct {" +
				field(1, "Float", "float64") +
				field(1, "Neg", "int") +
				field(1, "Pos", "int") +
				"\n}",
		},
	}

	trans := NewTransformer()
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			sn := "Out"
			if tt.structName != "" {
				sn = tt.structName
			}

			result, err := trans.Transform(sn, tt.input)
			assertEqualErr(t, tt.err, err)
			assertEqual(t, tt.output, result)
		})
	}
}

func assertEqualErr(t *testing.T, expected, actual error) {
	t.Helper()
	if expected == nil && actual == nil {
		return
	}

	if expected == nil || actual == nil {
		t.Errorf("expected: %v, got: %v", expected, actual)
		return
	}

	if !errors.Is(actual, expected) {
		t.Errorf("expected error: %v, got: %v", expected, actual)
	}
}

func assertEqual[T any](t *testing.T, expected, actual T) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected: %v, got: %v\n", expected, actual)
	}
}
