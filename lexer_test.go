package json2go

import "testing"

func TestLexer_Next(t *testing.T) {
	input := `{"name": "John", "age": 30, "active": true, "score": 3.14, "nothing": null}`
	tests := []struct {
		t TokenType
		l string
	}{
		{LBRACE, "{"},
		{STRING, "name"},
		{COLON, ":"},
		{INDENT, " "},
		{STRING, "John"},
		{COMMA, ","},
		{INDENT, " "},
		{STRING, "age"},
		{COLON, ":"},
		{INDENT, " "},
		{NUMBER, "30"},
		{COMMA, ","},
		{INDENT, " "},
		{STRING, "active"},
		{COLON, ":"},
		{INDENT, " "},
		{BOOL, "true"},
		{COMMA, ","},
		{INDENT, " "},
		{STRING, "score"},
		{COLON, ":"},
		{INDENT, " "},
		{DECIMAL, "3.14"},
		{COMMA, ","},
		{INDENT, " "},
		{STRING, "nothing"},
		{COLON, ":"},
		{INDENT, " "},
		{NULL, "null"},
		{RBRACE, "}"},
		{EOF, ""},
	}

	l := NewLexer([]byte(input))
	for i, tt := range tests {
		tok := l.Next()
		if tok.Type != tt.t {
			t.Errorf("tests[%d] - wrong token type. expected=%q, got=%q (literal=%q)", i, tt.t, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.l {
			t.Errorf("tests[%d] - wrong literal. expected=%q, got=%q", i, tt.l, tok.Literal)
		}
	}
}

func TestLexer_comments(t *testing.T) {
	input := `{
		// line comment
		"key": /* block comment */ "value"
	}`
	tests := []struct {
		t TokenType
		l string
	}{
		{LBRACE, "{"},
		{NEWLINE, "\n"},
		{INDENT, "\t\t"},
		{COMMENTLINE, " line comment"},
		{NEWLINE, "\n"},
		{INDENT, "\t\t"},
		{STRING, "key"},
		{COLON, ":"},
		{INDENT, " "},
		{COMMENTBLOCK, " block comment "},
		{INDENT, " "},
		{STRING, "value"},
		{NEWLINE, "\n"},
		{INDENT, "\t"},
		{RBRACE, "}"},
		{EOF, ""},
	}

	l := NewLexer([]byte(input))
	for i, tt := range tests {
		tok := l.Next()
		if tok.Type != tt.t {
			t.Errorf("tests[%d] - wrong token type. expected=%q, got=%q (literal=%q)", i, tt.t, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.l {
			t.Errorf("tests[%d] - wrong literal. expected=%q, got=%q", i, tt.l, tok.Literal)
		}
	}
}

func TestLexer_numbers(t *testing.T) {
	input := `[0, -1, 42, 3.14, -0.5, 1e10, 2.5E-3]`
	tests := []struct {
		t TokenType
		l string
	}{
		{LBRACKET, "["},
		{NUMBER, "0"},
		{COMMA, ","},
		{INDENT, " "},
		{NUMBER, "-1"},
		{COMMA, ","},
		{INDENT, " "},
		{NUMBER, "42"},
		{COMMA, ","},
		{INDENT, " "},
		{DECIMAL, "3.14"},
		{COMMA, ","},
		{INDENT, " "},
		{DECIMAL, "-0.5"},
		{COMMA, ","},
		{INDENT, " "},
		{DECIMAL, "1e10"},
		{COMMA, ","},
		{INDENT, " "},
		{DECIMAL, "2.5E-3"},
		{RBRACKET, "]"},
		{EOF, ""},
	}

	l := NewLexer([]byte(input))
	for i, tt := range tests {
		tok := l.Next()
		if tok.Type != tt.t {
			t.Errorf("tests[%d] - wrong token type. expected=%q, got=%q (literal=%q)", i, tt.t, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.l {
			t.Errorf("tests[%d] - wrong literal. expected=%q, got=%q", i, tt.l, tok.Literal)
		}
	}
}

func TestLexer_illegal(t *testing.T) {
	tests := []struct{ input, l string }{
		{`"unterminated`, "unterminated string"},
		{`/* unterminated`, "unterminated block comment"},
		{`01`, "leading zero in number"},
		{`-`, "invalid number"},
		{`1.`, "expected digit after decimal point"},
		{`1e`, "expected digit in exponent"},
	}
	for i, tt := range tests {
		l := NewLexer([]byte(tt.input))
		tok := l.Next()
		if tok.Type != ILLEGAL {
			t.Errorf("tests[%d] - expected ILLEGAL, got=%q (literal=%q)", i, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.l {
			t.Errorf("tests[%d] - wrong error literal. expected=%q, got=%q", i, tt.l, tok.Literal)
		}
	}
}
