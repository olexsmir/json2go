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

func TestTransform(t *testing.T) {
	tests := map[string]struct {
		input      string
		check      func(t *testing.T, result string)
		structName string
		err        error
	}{
		"simple object": {
			input: `{"name": "Olex", "active": true, "age": 420}`,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "type Out struct") {
					t.Errorf("missing Out struct")
				}
				if !strings.Contains(result, "Name string `json:\"name\"`") {
					t.Errorf("missing Name field")
				}
				if !strings.Contains(result, "Active bool `json:\"active\"`") {
					t.Errorf("missing Active field")
				}
				if !strings.Contains(result, "Age int `json:\"age\"`") {
					t.Errorf("missing Age field")
				}
			},
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
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "FirstName string `json:\"first_name\"`") {
					t.Errorf("missing FirstName field")
				}
				if !strings.Contains(result, "LastName string `json:\"last_name\"`") {
					t.Errorf("missing LastName field")
				}
			},
		},
		"nested object and array": {
			input: `{"user": {"name": "Alice", "score": 95.5}, "tags": ["go", "json"]}`,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "type Out struct") {
					t.Errorf("missing Out struct")
				}
				if !strings.Contains(result, "Tags []string") {
					t.Errorf("missing Tags field")
				}
				if !strings.Contains(result, "User struct {") {
					t.Errorf("missing inline User struct")
				}
				if !strings.Contains(result, "Name string `json:\"name\"`") {
					t.Errorf("missing Name field in inline struct")
				}
			},
		},
		"empty nested object": {
			input: `{"user": {}}`,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "type Out struct") {
					t.Errorf("missing Out struct")
				}
				if !strings.Contains(result, "User struct {") {
					t.Errorf("missing inline User struct")
				}
			},
		},
		"array of object": {
			input: `[{"name": "John"}, {"name": "Jane"}]`,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "type Out []struct {") {
					t.Errorf("missing Out array type with inline struct, got: %s", result)
				}
				if !strings.Contains(result, "Name string `json:\"name\"`") {
					t.Errorf("missing Name field")
				}
			},
		},
		"empty array": {
			input: `{"items": []}`,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Items []any `json:\"items\"`") {
					t.Errorf("missing Items field")
				}
			},
		},
		"null": {
			input: `{"item": null}`,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Item any `json:\"item\"`") {
					t.Errorf("missing Item field")
				}
			},
		},
		"numbers": {
			input: `{"pos": 123, "neg": -321, "float": 420.69}`,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Pos int `json:\"pos\"`") {
					t.Errorf("missing Pos field")
				}
				if !strings.Contains(result, "Neg int `json:\"neg\"`") {
					t.Errorf("missing Neg field")
				}
				if !strings.Contains(result, "Float float64 `json:\"float\"`") {
					t.Errorf("missing Float field")
				}
			},
		},
	}

	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			sn := "Out"
			if tt.structName != "" {
				sn = tt.structName
			}

			result, err := Transform(sn, tt.input, true)
			assertEqualErr(t, tt.err, err)
			if tt.check != nil {
				tt.check(t, result)
			}
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
