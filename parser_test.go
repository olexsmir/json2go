package json2go

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := map[string]struct {
		inp      string
		expected Value
		err      bool
	}{
		"string value": {
			inp:      `"hello"`,
			expected: Value{Kind: StringValue, Str: "hello"},
		},
		"integer value":    {inp: `42`, expected: Value{Kind: NumberValue, Int: 42}},
		"negative integer": {inp: `-42`, expected: Value{Kind: NumberValue, Int: -42}},
		"decimal value":    {inp: `3.14`, expected: Value{Kind: DecimalValue, Float: 3.14}},
		"bool true":        {inp: `true`, expected: Value{Kind: BoolValue, Bool: true}},
		"bool false":       {inp: `false`, expected: Value{Kind: BoolValue, Bool: false}},
		"null":             {inp: `null`, expected: Value{Kind: NullValue}},
		"empty object":     {inp: `{}`, expected: Value{Kind: ObjectValue}},
		"empty array":      {inp: `[]`, expected: Value{Kind: ArrayValue}},
		"flat object": {
			inp: `{"name": "John", "age": 30, "active": true}`,
			expected: Value{Kind: ObjectValue, Object: []Field{
				{"name", Value{Kind: StringValue, Str: "John"}},
				{"age", Value{Kind: NumberValue, Int: 30}},
				{"active", Value{Kind: BoolValue, Bool: true}},
			}},
		},
		"nested object": {
			inp: `{"user": {"name": "John", "age": 30}}`,
			expected: Value{Kind: ObjectValue, Object: []Field{
				{"user", Value{Kind: ObjectValue, Object: []Field{
					{"name", Value{Kind: StringValue, Str: "John"}},
					{"age", Value{Kind: NumberValue, Int: 30}},
				}}},
			}},
		},
		"array of numbers": {
			inp: `[1, 2, 3]`,
			expected: Value{Kind: ArrayValue, Array: []Value{
				{Kind: NumberValue, Int: 1},
				{Kind: NumberValue, Int: 2},
				{Kind: NumberValue, Int: 3},
			}},
		},
		"array of objects": {
			inp: `[{"a": 1}, {"a": 2}]`,
			expected: Value{Kind: ArrayValue, Array: []Value{
				{Kind: ObjectValue, Object: []Field{{"a", Value{Kind: NumberValue, Int: 1}}}},
				{Kind: ObjectValue, Object: []Field{{"a", Value{Kind: NumberValue, Int: 2}}}},
			}},
		},
		"object with line comment": {
			inp: `{
				// this is a comment
				"key": "value"
			}`,
			expected: Value{Kind: ObjectValue, Object: []Field{
				{"key", Value{Kind: StringValue, Str: "value"}},
			}},
		},
		"object with block comment": {
			inp: `{"key": /* comment */ "value"}`,
			expected: Value{Kind: ObjectValue, Object: []Field{
				{"key", Value{Kind: StringValue, Str: "value"}},
			}},
		},
		"trailing comma in object": {
			inp: `{"key": "value",}`,
			expected: Value{Kind: ObjectValue, Object: []Field{
				{"key", Value{Kind: StringValue, Str: "value"}},
			}},
		},
		"trailing comma in array": {
			inp: `[1, 2, 3,]`,
			expected: Value{Kind: ArrayValue, Array: []Value{
				{Kind: NumberValue, Int: 1},
				{Kind: NumberValue, Int: 2},
				{Kind: NumberValue, Int: 3},
			}},
		},
		"unterminated object": {inp: `{"key": "value"`, err: true},
		"unterminated array":  {inp: `[1, 2`, err: true},
		"missing colon":       {inp: `{"key" "value"}`, err: true},
		"non-string key":      {inp: `{42: "value"}`, err: true},
		"trailing content":    {inp: `{} 1`, err: true},
	}
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			l := NewLexer([]byte(tt.inp))
			p := NewParser(l)
			got, err := p.Parse()
			if tt.err {
				if err == nil {
					t.Errorf("expected error, got nil (value=%+v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("wrong value\nexpected: %+v\ngot:      %+v", tt.expected, got)
			}
		})
	}
}

// ensures the parser handles all valid json that encoding/json accepts.
func FuzzParser(f *testing.F) {
	f.Add([]byte("null"))
	f.Add([]byte("true"))
	f.Add([]byte("false"))
	f.Add([]byte("0"))
	f.Add([]byte("1.5"))
	f.Add([]byte("-42"))
	f.Add([]byte("1e10"))
	f.Add([]byte("1.5e-5"))
	f.Add([]byte(`""`))
	f.Add([]byte(`"hello"`))
	f.Add([]byte(`"hello\nworld"`))
	f.Add([]byte(`"hello\\nworld"`))
	f.Add([]byte(`"unicode\u0041"`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`[1,2,3]`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"a":1}`))
	f.Add([]byte(`{"a":1,"b":2}`))
	f.Add([]byte(`{"nested":{"x":1}}`))
	f.Add([]byte(`[{"a":1}]`))
	f.Add([]byte(`[1,[2,[3,[4]]]]`))
	f.Add([]byte(`{"a":"b","c":{"d":"e"}}`))
	f.Add([]byte(`[null,true,false]`))
	f.Add([]byte(`{"unicode":"🔥"}`))
	f.Add([]byte(`{ "a" : 1 }`))
	f.Add([]byte("{\n\t\"a\": 1\n}"))
	f.Add([]byte(`{"a":1/*comment*/,"b":2}`))
	f.Add([]byte(`{"a":1//comment
}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var any interface{}
		stdErr := json.Unmarshal(data, &any)

		lexer := NewLexer(data)
		parser := NewParser(lexer)
		_, parseErr := parser.Parse()

		// If encoding/json accepts it, our parser should too
		if stdErr == nil && parseErr != nil {
			t.Fatalf("encoding/json accepted but our parser rejected: %s\nerror: %v", string(data), parseErr)
		}

		// If our parser succeeded, transpiler should succeed too
		if parseErr == nil {
			_, transpileErr := Transform("Test", string(data), true)
			if transpileErr != nil {
				t.Fatalf("parser succeeded but transpiler failed: %s\nerror: %v", string(data), transpileErr)
			}
		}
	})
}
