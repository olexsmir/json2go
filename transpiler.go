package json2go

import (
	"strings"
	"unicode"
)

// Transpiler transpiles AST [Value] to Go type definitions.
type Transpiler struct{}

func NewTranspiler() *Transpiler { return &Transpiler{} }

// Transpile converts a [Value] AST to Go type definitions.
func (t *Transpiler) Transpile(structName string, v Value, includeTags bool) (string, error) {
	var buf strings.Builder
	buf.WriteString("type ")
	buf.WriteString(structName)

	switch v.Kind {
	case ArrayValue:
		buf.WriteString(" [")
		if len(v.Array) == 0 {
			buf.WriteString("]any")
		} else {
			buf.WriteByte(']')
			t.writeInlineType(&buf, structName+"Item", v.Array[0], includeTags)
		}

	case ObjectValue:
		buf.WriteByte(' ')
		t.writeInlineStruct(&buf, structName, v.Object, includeTags)

	default:
		buf.WriteByte(' ')
		t.writeScalarType(&buf, v)
	}

	return buf.String(), nil
}

func (t *Transpiler) writeInlineType(buf *strings.Builder, name string, v Value, includeTags bool) {
	switch v.Kind {
	case ObjectValue:
		t.writeInlineStruct(buf, name, v.Object, includeTags)

	case ArrayValue:
		buf.WriteByte('[')
		if len(v.Array) == 0 {
			buf.WriteString("]any")
		} else {
			buf.WriteByte(']')
			t.writeInlineType(buf, name+"Item", v.Array[0], includeTags)
		}

	default:
		t.writeScalarType(buf, v)
	}
}

func (t *Transpiler) writeInlineStruct(buf *strings.Builder, name string, fields []Field, includeTags bool) {
	buf.WriteString("struct {\n")
	for _, f := range fields {
		fieldName := t.sanitizeFieldName(f.K)
		buf.WriteByte('\t')
		buf.WriteString(fieldName)
		buf.WriteByte(' ')
		t.writeInlineType(buf, name+fieldName, f.V, includeTags)
		if includeTags {
			buf.WriteString(" `json:\"")
			buf.WriteString(f.K)
			buf.WriteString("\"`")
		}
		buf.WriteByte('\n')
	}
	buf.WriteByte('}')
}

func (t *Transpiler) writeScalarType(buf *strings.Builder, v Value) {
	switch v.Kind {
	case StringValue:
		buf.WriteString("string")
	case NumberValue:
		buf.WriteString("int")
	case DecimalValue:
		buf.WriteString("float64")
	case BoolValue:
		buf.WriteString("bool")
	default:
		buf.WriteString("any")
	}
}

func (t *Transpiler) sanitizeFieldName(jsonKey string) string {
	if jsonKey == "" {
		return "Field"
	}

	var result strings.Builder
	result.Grow(len(jsonKey))

	capitalize := true
	for _, r := range jsonKey {
		if r == '_' {
			capitalize = true
			continue
		}
		if capitalize {
			result.WriteRune(unicode.ToUpper(r))
			capitalize = false
		} else {
			result.WriteRune(r)
		}
	}

	name := result.String()
	if name != "" && isValidIdentifier(name) {
		return name
	}

	return t.sanitizeInvalidIdentifier(jsonKey)
}

func (t *Transpiler) sanitizeInvalidIdentifier(jsonKey string) string {
	var result strings.Builder
	for i, r := range jsonKey {
		if unicode.IsLetter(r) || r == '_' || (i > 0 && unicode.IsDigit(r)) {
			result.WriteRune(r)
		} else if i > 0 && r == '-' {
			result.WriteRune('_')
		}
	}

	name := result.String()
	if name == "" || (!unicode.IsLetter(rune(name[0])) && name[0] != '_') {
		var b strings.Builder
		b.Grow(len(name) + 1)
		b.WriteByte('F')
		b.WriteString(name)
		return b.String()
	}

	return name
}
