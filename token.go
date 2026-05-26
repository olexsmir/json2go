package json2go

type TokenType int

type Token struct {
	Type    TokenType
	Literal string
}

//go:generate go tool stringer -output token_type_string.go -type=TokenType
const (
	EOF TokenType = iota
	ILLEGAL

	NEWLINE // \n
	INDENT  // leading whitespace

	NUMBER  // 420
	DECIMAL // 69.420
	STRING  // "quote string"
	BOOL    // "true" "false"
	NULL    // "null"

	COLON    // :
	COMMA    // ,
	LBRACE   // {
	RBRACE   // }
	LBRACKET // [
	RBRACKET // ]

	COMMENTLINE  // // ...
	COMMENTBLOCK // /* ... */
)
