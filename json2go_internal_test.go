package json2go

import "testing"

func TestTransformer_GetGoType(t *testing.T) {
	tests := map[string]struct {
		value     any
		fieldName string
		output    string
	}{
		"struct": {
			value: map[string]any{
				"username": "user-ovich",
				"age":      float64(20),
			},
			output: "struct {" +
				field(1, "Age", "int") +
				field(1, "Username", "string") +
				"\n}",
		},
		"empty slice": {
			value:  make([]any, 0),
			output: "[]any",
		},
		"slice of ints": {
			value:  []any{float64(3), float64(123)},
			output: "[]int",
		},
		"slice of floats": {
			value:  []any{float64(3.4), float64(123.3)},
			output: "[]float64",
		},
		"slice of strings": {
			value:  []any{"asdf", "jalkjsd"},
			output: "[]string",
		},
		"slice of bool": {
			value:  []any{false, true, false},
			output: "[]bool",
		},
		"int": {
			value:  float64(1233),
			output: "int",
		},
		"float64": {
			value:  float64(1233.23),
			output: "float64",
		},
		"bool": {
			value:  false,
			output: "bool",
		},
		"any": {
			value:  nil,
			output: "any",
		},
	}

	trans := NewTransformer()
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			t.Parallel()

			fieldName := "field"
			if tt.fieldName != "" {
				fieldName = tt.fieldName
			}

			res := trans.getGoType(fieldName, tt.value)
			assertEqual(t, tt.output, res)
		})
	}
}

func TestTransformer_buildStruct(t *testing.T) {
	tests := map[string]struct {
		input  map[string]any
		output string
	}{
		"simple struct": {
			input: map[string]any{
				// only one value, because of the inconsistent ordering of maps
				"active": true,
			},
			output: "struct {" +
				field(1, "Active", "bool", "active") +
				"\n}",
		},
		"with no named field": {
			input: map[string]any{"": "user"},
			output: "struct {" +
				field(1, "NotNamedField", "string", "NotNamedField") +
				"\n}",
		},
	}

	trans := NewTransformer()
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			t.Parallel()

			res := trans.buildStruct(tt.input)
			assertEqual(t, tt.output, res)
		})
	}
}

func TestTransformer_getTypeAnnotation(t *testing.T) {
	c := "type Typeich "
	tests := map[string]struct {
		input  any
		output string
	}{
		"struct": {
			input: map[string]any{"field": false},
			output: c + "struct {" +
				field(1, "Field", "bool") + "\n}",
		},
		"slice": {
			input:  []any{"asdf", "jkl;"},
			output: c + "[]string",
		},
		"empty slice": {
			input:  make([]any, 0),
			output: c + "[]any",
		},
		"string": {
			input:  "asdf",
			output: c + "string",
		},
		"int": {
			input:  float64(123),
			output: c + "int",
		},
		"float64": {
			input:  float64(123.69),
			output: c + "float64",
		},
		"bool": {
			input:  true,
			output: c + "bool",
		},
		"any": {
			input:  nil,
			output: c + "any",
		},
	}

	trans := NewTransformer()
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			t.Parallel()

			res := trans.getTypeAnnotation("Typeich", tt.input)
			assertEqual(t, tt.output, res)
		})
	}
}

func TestTransformer_toGoFieldName(t *testing.T) {
	tests := map[string]string{
		"input":         "Input",
		"Input":         "Input",
		"long_name":     "LongName",
		"a_lot_of_____": "ALotOf",
		"__name":        "Name",
	}

	trans := NewTransformer()
	for input, output := range tests {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			res := trans.toGoFieldName(input)
			assertEqual(t, output, res)
		})
	}
}

func TestMapToStructInput(t *testing.T) {
	inp := map[string]any{
		"field1": nil,
		"field2": true,
		"a":      123,
		"user":   map[string]any{},
	}

	assertEqual(t, mapToStructInput(inp), []structInput{
		{"a", 123},
		{"field1", nil},
		{"field2", true},
		{"user", map[string]any{}},
	})
}
