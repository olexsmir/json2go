package json2go

import (
	"strings"
	"unicode"
)

// Transpiler transpiles AST [Value] to Go type definitions.
type Transpiler struct{}

func NewTranspiler() *Transpiler { return &Transpiler{} }

// Transpile converts a [Value] AST to Go type definitions.
// Nested types are emitted after the parent struct.
func (t *Transpiler) Transpile(structName string, v Value, includeTags bool) (string, error) {
	var buf strings.Builder
	var nested strings.Builder

	switch v.Kind {
	case ArrayValue:
		if len(v.Array) == 0 {
			buf.WriteString("type ")
			buf.WriteString(structName)
			buf.WriteString(" []any")
		} else {
			itemType := t.writeWithNested(&buf, &nested, structName+"Item", v.Array[0], includeTags)
			buf.WriteString("type ")
			buf.WriteString(structName)
			buf.WriteString(" []")
			buf.WriteString(itemType)
		}

	case ObjectValue:
		t.writeStructDef(&buf, &nested, structName, v.Object, includeTags)

	default:
		scalarType := t.inferType(structName, v)
		buf.WriteString("type ")
		buf.WriteString(structName)
		buf.WriteByte(' ')
		buf.WriteString(scalarType)
	}

	if nested.Len() > 0 && buf.Len() > 0 {
		buf.WriteString("\n\n")
		buf.WriteString(nested.String())
	}
	return buf.String(), nil
}

func (t *Transpiler) writeWithNested(buf, nested *strings.Builder, name string, v Value, includeTags bool) string {
	switch v.Kind {
	case ObjectValue:
		t.writeStructDef(nested, nested, name, v.Object, includeTags)
		return name

	case ArrayValue:
		if len(v.Array) == 0 {
			return "[]any"
		}
		itemType := t.writeWithNested(buf, nested, name+"Item", v.Array[0], includeTags)
		return "[]" + itemType
	}
	return t.inferType(name, v)
}

func (t *Transpiler) writeStructDef(buf, nested *strings.Builder, name string, fields []Field, includeTags bool) {
	if buf.Len() > 0 {
		buf.WriteString("\n\n")
	}

	buf.WriteString("type ")
	buf.WriteString(name)
	buf.WriteString(" struct {\n")
	for _, f := range fields {
		fieldName := t.sanitizeFieldName(f.K)
		fieldType := t.writeWithNested(buf, nested, name+fieldName, f.V, includeTags)
		if includeTags {
			buf.WriteByte('\t')
			buf.WriteString(fieldName)
			buf.WriteByte(' ')
			buf.WriteString(fieldType)
			buf.WriteString(" `json:\"")
			buf.WriteString(f.K)
			buf.WriteString("\"`\n")
		} else {
			buf.WriteByte('\t')
			buf.WriteString(fieldName)
			buf.WriteByte(' ')
			buf.WriteString(fieldType)
			buf.WriteByte('\n')
		}
	}
	buf.WriteString("}")
}

func (t *Transpiler) inferType(name string, v Value) string {
	switch v.Kind {
	case ObjectValue:
		return name
	case ArrayValue:
		if len(v.Array) == 0 {
			return "[]any"
		}
		itemType := t.inferType(name+"Item", v.Array[0])
		return "[]" + itemType
	case StringValue:
		return "string"
	case NumberValue:
		return "int"
	case DecimalValue:
		return "float64"
	case BoolValue:
		return "bool"
	case NullValue:
		return "any"
	default:
		return "any"
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
