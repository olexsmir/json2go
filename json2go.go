package json2go

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var (
	ErrInvalidJSON       = errors.New("invalid json")
	ErrInvalidStructName = errors.New("invalid struct name")
)

type Transformer struct {
	structName    string
	currentIndent int
}

func NewTransformer() *Transformer {
	return &Transformer{}
}

// Transform transforms provided json string into go type annotation
func (t *Transformer) Transform(structName, jsonStr string) (string, error) {
	if !identRe.MatchString(structName) {
		return "", ErrInvalidStructName
	}

	t.structName = structName
	t.currentIndent = 0

	var input any
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		return "", errors.Join(ErrInvalidJSON, err)
	}

	type_ := t.getTypeAnnotation(structName, input)
	return type_, nil
}

func (t *Transformer) getTypeAnnotation(typeName string, input any) string {
	switch v := input.(type) {
	case map[string]any:
		return fmt.Sprintf("type %s %s", typeName, t.buildStruct(v))

	case []any:
		if len(v) == 0 {
			return fmt.Sprintf("type %s []any", typeName)
		}

		type_ := t.getGoType(typeName+"Item", v[0])
		return fmt.Sprintf("type %s []%s", typeName, type_)

	case string:
		return fmt.Sprintf("type %s string", typeName)

	case float64:
		if float64(int(v)) == v {
			return fmt.Sprintf("type %s int", typeName)
		}
		return fmt.Sprintf("type %s float64", typeName)

	case bool:
		return fmt.Sprintf("type %s bool", typeName)

	default:
		return fmt.Sprintf("type %s any", typeName)

	}
}

func (t *Transformer) buildStruct(input map[string]any) string {
	var fields strings.Builder
	for _, f := range mapToStructInput(input) {
		fieldName := t.toGoFieldName(f.field)
		if fieldName == "" {
			fieldName = "NotNamedField"
			f.field = "NotNamedField"
		}

		// increase indentation in case of building new struct
		t.currentIndent++
		fieldType := t.getGoType(fieldName, f.type_)
		t.currentIndent--

		jsonTag := fmt.Sprintf("`json:\"%s\"`", f.field)

		indent := strings.Repeat("\t", t.currentIndent+1)
		fields.WriteString(fmt.Sprintf(
			"%s%s %s %s\n",
			indent,
			fieldName,
			fieldType,
			jsonTag,
		))
	}

	return fmt.Sprintf("struct {\n%s%s}",
		fields.String(),
		strings.Repeat("\t", t.currentIndent))
}

func (t *Transformer) getGoType(fieldName string, value any) string {
	switch v := value.(type) {
	case map[string]any:
		return t.buildStruct(v)

	case []any:
		if len(v) == 0 {
			return "[]any"
		}

		type_ := t.getGoType(fieldName, v[0])
		return "[]" + type_

	case float64:
		if float64(int(v)) == v {
			return "int"
		}
		return "float64"

	case string:
		return "string"

	case bool:
		return "bool"

	default:
		return "any"
	}
}

func (t *Transformer) toGoFieldName(jsonField string) string {
	parts := strings.Split(jsonField, "_")

	var result strings.Builder
	for _, part := range parts {
		if part != "" {
			if len(part) > 0 {
				result.WriteString(strings.ToUpper(part[:1]) + part[1:])
			}
		}
	}

	return result.String()
}

type structInput struct {
	field string
	type_ any
}

func mapToStructInput(input map[string]any) []structInput {
	res := make([]structInput, 0, len(input))
	for k, v := range input {
		res = append(res, structInput{k, v})
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].field < res[j].field
	})

	return res
}
