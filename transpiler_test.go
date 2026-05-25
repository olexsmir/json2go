package json2go

import (
	"strings"
	"testing"
)

func TestTranspiler_Transpile_WithTags(t *testing.T) {
	tests := map[string]struct {
		v     Value
		name  string
		check func(t *testing.T, result string)
	}{
		"simple object": {
			name: "User",
			v: Value{
				Kind: ObjectValue,
				Object: []Field{
					{K: "name", V: Value{Kind: StringValue, Str: "John"}},
					{K: "age", V: Value{Kind: NumberValue, Int: 30}},
				},
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "type User struct") {
					t.Errorf("missing User struct")
				}
				if !strings.Contains(result, "Name string `json:\"name\"`") {
					t.Errorf("missing Name field")
				}
				if !strings.Contains(result, "Age int `json:\"age\"`") {
					t.Errorf("missing Age field")
				}
			},
		},
		"nested object": {
			name: "Response",
			v: Value{
				Kind: ObjectValue,
				Object: []Field{{
					K: "user",
					V: Value{
						Kind: ObjectValue,
						Object: []Field{
							{K: "name", V: Value{Kind: StringValue, Str: "Alice"}},
							{K: "active", V: Value{Kind: BoolValue, Bool: true}},
						},
					},
				}},
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "type Response struct") {
					t.Errorf("missing Response struct")
				}
				if !strings.Contains(result, "type ResponseUser struct") {
					t.Errorf("missing ResponseUser struct")
				}
				if !strings.Contains(result, "User ResponseUser `json:\"user\"`") {
					t.Errorf("missing User field")
				}
			},
		},
		"array of scalars": {
			name: "Tags",
			v: Value{
				Kind:  ArrayValue,
				Array: []Value{{Kind: StringValue, Str: "go"}},
			},
			check: func(t *testing.T, result string) {
				if result != "type Tags []string" {
					t.Errorf("expected scalar array type, got: %s", result)
				}
			},
		},
		"array of objects": {
			name: "Users",
			v: Value{
				Kind: ArrayValue,
				Array: []Value{{
					Kind: ObjectValue,
					Object: []Field{
						{K: "id", V: Value{Kind: NumberValue, Int: 1}},
						{K: "name", V: Value{Kind: StringValue, Str: "Bob"}},
					},
				}},
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "type UsersItem struct") {
					t.Errorf("missing UsersItem struct in: %s", result)
				}
				if !strings.Contains(result, "Id int `json:\"id\"`") {
					t.Errorf("missing Id field")
				}
			},
		},
		"snake_case fields": {
			name: "Data",
			v: Value{
				Kind: ObjectValue,
				Object: []Field{
					{K: "first_name", V: Value{Kind: StringValue, Str: "Jane"}},
					{K: "last_name", V: Value{Kind: StringValue, Str: "Doe"}},
				},
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "FirstName string `json:\"first_name\"`") {
					t.Errorf("missing FirstName field")
				}
				if !strings.Contains(result, "LastName string `json:\"last_name\"`") {
					t.Errorf("missing LastName field")
				}
			},
		},
	}
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			result, err := NewTranspiler().Transpile(tt.name, tt.v, true)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestTranspiler_Transpile_WithoutTags(t *testing.T) {
	tests := map[string]struct {
		v     Value
		name  string
		check func(t *testing.T, result string)
	}{
		"simple object": {
			name: "User",
			v: Value{
				Kind: ObjectValue,
				Object: []Field{
					{K: "name", V: Value{Kind: StringValue, Str: "John"}},
					{K: "age", V: Value{Kind: NumberValue, Int: 30}},
				},
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "type User struct") {
					t.Errorf("missing User struct")
				}
				if !strings.Contains(result, "Name string\n") {
					t.Errorf("missing Name field without tag")
				}
				if !strings.Contains(result, "Age int\n") {
					t.Errorf("missing Age field without tag")
				}
				if strings.Contains(result, "`json:") {
					t.Errorf("should not have json tags")
				}
			},
		},
		"nested object": {
			name: "Response",
			v: Value{
				Kind: ObjectValue,
				Object: []Field{{
					K: "user",
					V: Value{
						Kind: ObjectValue,
						Object: []Field{
							{K: "name", V: Value{Kind: StringValue, Str: "Alice"}},
							{K: "active", V: Value{Kind: BoolValue, Bool: true}},
						},
					},
				}},
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "type Response struct") {
					t.Errorf("missing Response struct")
				}
				if !strings.Contains(result, "type ResponseUser struct") {
					t.Errorf("missing ResponseUser struct")
				}
				if strings.Contains(result, "`json:") {
					t.Errorf("should not have json tags")
				}
			},
		},
		"snake_case fields preserved": {
			name: "Data",
			v: Value{
				Kind: ObjectValue,
				Object: []Field{
					{K: "first_name", V: Value{Kind: StringValue, Str: "Jane"}},
					{K: "last_name", V: Value{Kind: StringValue, Str: "Doe"}},
				},
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "FirstName string\n") {
					t.Errorf("missing FirstName field")
				}
				if !strings.Contains(result, "LastName string\n") {
					t.Errorf("missing LastName field")
				}
			},
		},
	}
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			result, err := NewTranspiler().Transpile(tt.name, tt.v, false)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestTranspiler_ParserIntegration(t *testing.T) {
	tests := map[string]struct {
		input       string
		structName  string
		includeTags bool
		expectTags  bool
	}{
		"parse and transpile with tags": {
			input:       `{"user_name": "alice", "user_age": 25}`,
			structName:  "Profile",
			includeTags: true,
			expectTags:  true,
		},
		"parse and transpile without tags": {
			input:       `{"user_name": "bob", "user_age": 30}`,
			structName:  "Account",
			includeTags: false,
			expectTags:  false,
		},
	}
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			lexer := NewLexer([]byte(tt.input))
			parser := NewParser(lexer)
			v, err := parser.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			result, err := NewTranspiler().Transpile(tt.structName, v, tt.includeTags)
			if err != nil {
				t.Fatalf("transpile error: %v", err)
			}

			hasTag := strings.Contains(result, "`json:")
			if tt.expectTags != hasTag {
				t.Errorf("expected tags=%v, got=%v\n%s", tt.expectTags, hasTag, result)
			}

			if !strings.Contains(result, "type "+tt.structName) {
				t.Errorf("struct name %q not found in output", tt.structName)
			}
		})
	}
}
