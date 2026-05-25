package json2go

import (
	"fmt"
	"strconv"
)

type Parser struct {
	lexer *Lexer
	cur   Token
	peek  Token
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{lexer: l}
	p.advance() // populate .peek
	p.advance() // populate .cur
	return p
}

// Parse starts parsing and returns the root Value AST.
// Expects well-formed JSON with a single top-level value.
// Returns an error if the JSON is malformed or has unexpected tokens after the main value.
func (p *Parser) Parse() (Value, error) {
	p.skipNoise()
	v, err := p.parseValue()
	if err != nil {
		return Value{}, err
	}
	p.skipNoise()
	if !p.got(EOF) {
		return Value{}, fmt.Errorf("unexpected token after value: %q", p.cur.Literal)
	}
	return v, nil
}

func (p *Parser) parseValue() (Value, error) {
	p.skipNoise()
	switch p.cur.Type {
	case LBRACE:
		return p.parseObject()
	case LBRACKET:
		return p.parseArray()
	case STRING:
		v := Value{Kind: StringValue, Str: p.cur.Literal}
		p.advance()
		return v, nil
	case NUMBER:
		n, err := strconv.ParseInt(p.cur.Literal, 10, 64)
		if err != nil {
			f, ferr := strconv.ParseFloat(p.cur.Literal, 64)
			if ferr != nil {
				return Value{}, fmt.Errorf("invalid number: %w", err)
			}
			p.advance()
			return Value{Kind: DecimalValue, Float: f}, nil
		}
		p.advance()
		return Value{Kind: NumberValue, Int: n}, nil
	case DECIMAL:
		f, err := strconv.ParseFloat(p.cur.Literal, 64)
		if err != nil {
			return Value{}, fmt.Errorf("invalid decimal: %w", err)
		}
		p.advance()
		return Value{Kind: DecimalValue, Float: f}, nil
	case BOOL:
		v := Value{Kind: BoolValue, Bool: p.cur.Literal == "true"}
		p.advance()
		return v, nil
	case NULL:
		p.advance()
		return Value{Kind: NullValue}, nil
	default:
		return Value{}, fmt.Errorf("unexpected token %q (%q)", p.cur.Type, p.cur.Literal)
	}
}

func (p *Parser) parseObject() (Value, error) {
	p.advance()
	var fields []Field
	for {
		p.skipNoise()
		if p.got(RBRACE) {
			p.advance()
			return Value{Kind: ObjectValue, Object: fields}, nil
		}
		if p.got(EOF) {
			return Value{}, fmt.Errorf("unterminated object")
		}

		keyTok, err := p.expect(STRING)
		if err != nil {
			return Value{}, err
		}

		p.skipNoise()
		if _, cerr := p.expect(COLON); cerr != nil {
			return Value{}, cerr
		}

		val, err := p.parseValue()
		if err != nil {
			return Value{}, err
		}
		fields = append(fields, Field{keyTok.Literal, val})

		p.skipNoise()
		if p.got(COMMA) {
			p.advance()
		} else if !p.got(RBRACE) {
			_, err := p.expect(RBRACE)
			return Value{}, err
		}
	}
}

func (p *Parser) parseArray() (Value, error) {
	p.advance()
	var items []Value
	for {
		p.skipNoise()
		if p.got(RBRACKET) {
			p.advance()
			return Value{Kind: ArrayValue, Array: items}, nil
		}
		if p.got(EOF) {
			return Value{}, fmt.Errorf("unterminated array")
		}

		val, err := p.parseValue()
		if err != nil {
			return Value{}, err
		}
		items = append(items, val)

		p.skipNoise()
		if p.got(COMMA) {
			p.advance()
		} else if !p.got(RBRACKET) {
			_, err := p.expect(RBRACKET)
			return Value{}, err
		}
	}
}

func (p *Parser) got(kind TokenType) bool { return p.cur.Type == kind }
func (p *Parser) advance() Token {
	prev := p.cur
	p.cur = p.peek
	p.peek = p.lexer.Next()
	return prev
}

func (p *Parser) expect(kind TokenType) (Token, error) {
	if p.got(kind) {
		return p.advance(), nil
	}
	return p.cur, fmt.Errorf("expected %s, got %s", kind, p.cur.Type)
}

func (p *Parser) skipNoise() {
	for p.cur.Type == NEWLINE || p.cur.Type == INDENT ||
		p.cur.Type == COMMENTLINE || p.cur.Type == COMMENTBLOCK {
		p.advance()
	}
}
