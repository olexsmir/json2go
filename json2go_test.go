package json2go

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func field(name, type_ string, json_ ...string) string {
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
				field("Name", "string") +
				field("Active", "bool") +
				field("Age", "int") +
				"\n}\n",
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
				"\n}\n",
		},
		"nested object and array": {
			input: `{"user": {"name": "Alice", "score": 95.5}, "tags": ["go", "json"]}`,
			output: "type Out struct {" +
				field("User", "User") +
				field("Tags", "[]string") +
				"\n}\ntype User struct {" +
				field("Name", "string") +
				field("Score", "float64") +
				"\n}\n",
		},
		"empty nested object": {
			input: `{"user": {}}`,
			output: "type Out struct {" +
				field("User", "User") +
				"\n}\ntype User struct {\n}\n",
		},
		"array of object": {
			input: `[{"name": "John"}, {"name": "Jane"}]`,
			output: "type Out []OutItem" +
				"\ntype OutItem struct {" +
				field("Name", "string") +
				"\n}\n",
		},
		"empty array": {
			input: `{"items": []}`,
			output: `type Out struct {` +
				field("Items", "[]any") + "\n}\n",
		},
		"null": {
			input: `{"item": null}`,
			output: `type Out struct {` +
				field("Item", "any") + "\n}\n",
		},
		"numbers": {
			input: `{"pos": 123, "neg": -321, "float": 420.69}`,
			output: "type Out struct {" +
				field("Pos", "int") +
				field("Neg", "int") +
				field("Float", "float64") +
				"\n}\n",
		},
	}

	trans := NewTransformer()
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			result, err := trans.Transform("Out", tt.input)
			assertEqualErr(t, tt.err, err)

			lines := strings.Split(result, "\n")
			counts := make(map[string]int)
			for _, line := range lines {
				if !strings.Contains(line, "}") {
					counts[line]++
				}
			}

			for _, line := range lines {
				if counts[line] > 1 {
					t.Fatalf("found duplicate line: %s", line)
				}
			}
		})
	}
}

func assertEqualErr(t *testing.T, expected, actual error) {
	t.Helper()
	if (expected != nil || actual != nil) && errors.Is(expected, actual) {
		t.Errorf("expected: %v, got: %v\n", expected, actual)
	}
}
