package json2go

import (
	"unicode/utf8"
)

type Lexer struct {
	input  []byte
	ch     rune // current rune (0 == EOF)
	chSize int  // byte size of [ch]
	pos    int  // current byte offset (points at [ch])
	rpos   int  // next byte offset to read (one ahead of [pos])
	col    int  // current column (1-based)
	line   int  // current line (1-based)
}

func NewLexer(input []byte) *Lexer {
	l := &Lexer{input: input, line: 1}
	l.advance()
	if l.ch == '\uFEFF' { // start of the input
		l.advance()
	}
	return l
}

// Next returns the next token from the input.
// Returns EOF when input is exhausted.
func (l *Lexer) Next() Token {
	switch {
	case l.ch == 0:
		return Token{EOF, ""}
	case l.ch == '\n', l.ch == '\r':
		l.advance()
		return Token{NEWLINE, "\n"}
	case l.ch == ' ', l.ch == '\t':
		offset := l.pos
		for l.ch == ' ' || l.ch == '\t' {
			l.advance()
		}
		return Token{INDENT, string(l.input[offset:l.pos])}
	case l.ch == '/':
		return l.lexComment()
	case l.ch == '"':
		return l.lexString()
	case l.ch == ':':
		l.advance()
		return Token{COLON, ":"}
	case l.ch == ',':
		l.advance()
		return Token{COMMA, ","}
	case l.ch == '[':
		l.advance()
		return Token{LBRACKET, "["}
	case l.ch == ']':
		l.advance()
		return Token{RBRACKET, "]"}
	case l.ch == '{':
		l.advance()
		return Token{LBRACE, "{"}
	case l.ch == '}':
		l.advance()
		return Token{RBRACE, "}"}
	case l.isDigit(), l.ch == '-':
		return l.lexNumber()
	case l.isAlpha():
		offset := l.pos
		for l.isAlpha() {
			l.advance()
		}
		lit := string(l.input[offset:l.pos])
		kind := ILLEGAL
		if lit == "false" || lit == "true" {
			kind = BOOL
		} else if lit == "null" {
			kind = NULL
		}
		return Token{kind, lit}
	}
	ch := l.ch
	l.advance()
	return Token{ILLEGAL, string(ch)}
}

func (l *Lexer) lexString() Token {
	l.advance()
	offset := l.pos
	for {
		switch l.ch {
		default:
			l.advance()
		case 0, '\r', '\n':
			return Token{ILLEGAL, "unterminated string"}
		case '\\':
			l.advance() // consume '\'
			switch l.ch {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				l.advance()
			case 'u':
				l.advance()
				for range 4 { // expect exactly 4 hex digits
					if !l.isHex() {
						return Token{ILLEGAL, "invalid unicode escape"}
					}
					l.advance()
				}
			default:
				return Token{ILLEGAL, "invalid escape sequence"}
			}
		case '"':
			lit := string(l.input[offset:l.pos])
			l.advance() // consume closing '"'
			return Token{STRING, lit}
		}
	}
}

func (l *Lexer) lexNumber() Token {
	offset := l.pos

	if l.ch == '-' { // optional leading minus
		l.advance()
	}

	// integer part
	if l.ch == '0' {
		l.advance()
		if l.isDigit() { // leading zero must not be followed by another digit
			return Token{ILLEGAL, "leading zero in number"}
		}
	} else if l.isDigit() {
		for l.isDigit() {
			l.advance()
		}
	} else {
		return Token{ILLEGAL, "invalid number"}
	}

	kind := NUMBER
	if l.ch == '.' { // optional fractional part
		kind = DECIMAL
		l.advance()
		if !l.isDigit() {
			return Token{ILLEGAL, "expected digit after decimal point"}
		}
		for l.isDigit() {
			l.advance()
		}
	}

	if l.ch == 'e' || l.ch == 'E' { // optional exponent
		kind = DECIMAL
		l.advance()
		if l.ch == '+' || l.ch == '-' {
			l.advance()
		}
		if !l.isDigit() {
			return Token{ILLEGAL, "expected digit in exponent"}
		}
		for l.isDigit() {
			l.advance()
		}
	}

	return Token{kind, string(l.input[offset:l.pos])}
}

func (l *Lexer) lexComment() Token {
	l.advance()
	switch l.ch {
	default:
		return Token{ILLEGAL, "invalid comment"}
	case '/':
		l.advance()
		offset := l.pos
		for l.ch != 0 && l.ch != '\n' && l.ch != '\r' {
			l.advance()
		}
		return Token{COMMENTLINE, string(l.input[offset:l.pos])}
	case '*':
		l.advance()
		offset := l.pos
		for {
			if l.ch == 0 {
				return Token{ILLEGAL, "unterminated block comment"}
			}
			if l.ch == '*' {
				l.advance()
				if l.ch == '/' {
					end := l.pos - 1 // exclude the '*'
					l.advance()
					return Token{COMMENTBLOCK, string(l.input[offset:end])}
				}
			} else {
				l.advance()
			}
		}
	}
}

func (l *Lexer) advance() {
	if l.rpos >= len(l.input) {
		l.ch = 0
		l.chSize = 0
	} else {
		l.ch, l.chSize = utf8.DecodeRune(l.input[l.rpos:])
	}
	l.pos = l.rpos
	l.rpos += l.chSize
	if l.ch == '\n' || l.ch == '\r' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
}
func (l *Lexer) isDigit() bool { return l.ch >= '0' && l.ch <= '9' }
func (l *Lexer) isAlpha() bool {
	return (l.ch >= 'a' && l.ch <= 'z') || (l.ch >= 'A' && l.ch <= 'Z')
}

func (l *Lexer) isHex() bool {
	return (l.ch >= '0' && l.ch <= '9') ||
		(l.ch >= 'a' && l.ch <= 'f') || (l.ch >= 'A' && l.ch <= 'F')
}
