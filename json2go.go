package json2go

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidJSON = errors.New("invalid json")

type (
	types       struct{ name, def string }
	Transformer struct {
		structName string
		types      []types
	}
)

func NewTransformer() *Transformer {
	return &Transformer{}
}

// Transform ...
// todo: take io.Reader as input?
// todo: output as io.Writer?
// todo: validate provided structName
func (t *Transformer) Transform(structName, jsonStr string) (string, error) {
	t.structName = structName
	t.types = make([]types, 1)

	var input any
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		return "", errors.Join(ErrInvalidJSON, err)
	}

	var result strings.Builder

	// the "parent" type
	type_ := t.generateTypeAnnotation(structName, input)
	result.WriteString(type_)

	// nested types
	for _, t := range t.types {
		if t.name != structName {
			result.WriteString(t.def)
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

func (t *Transformer) generateTypeAnnotation(typeName string, input any) string {
	switch v := input.(type) {
	case map[string]any:
		return t.buildStruct(typeName, v)

	case []any:
		if len(v) == 0 {
			return fmt.Sprintf("type %s []any", t.structName)
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

	case nil:
		return fmt.Sprintf("type %s any", typeName)

	default:
		return fmt.Sprintf("type %s any", typeName)

	}
}

// todo: input shouldn't be map, to preserve it's order
func (t *Transformer) buildStruct(typeName string, input map[string]any) string {
	var fields strings.Builder
	for key, value := range input {
		fieldName := t.toGoFieldName(key)
		if fieldName == "" {
			fieldName = "Field"
		}

		fieldType := t.getGoType(fieldName, value)

		// todo: toggle json tags generation
		jsonTag := fmt.Sprintf("`json:\"%s\"`", key)

		// todo: figure out the indentation, since it might have nested struct
		fields.WriteString(fmt.Sprintf(
			"%s %s %s\n",
			fieldName,
			fieldType,
			jsonTag,
		))
	}

	structDef := fmt.Sprintf("type %s struct {\n%s}", typeName, fields.String())
	t.types = append(t.types, types{
		name: typeName,
		def:  structDef,
	})

	return structDef
}

func (t *Transformer) getGoType(fieldName string, value any) string {
	switch v := value.(type) {
	case map[string]any:
		typeName := t.toGoTypeName(fieldName)
		if !t.isTypeRecorded(typeName) {
			t.generateTypeAnnotation(typeName, v)
		}
		return typeName

	case []any:
		if len(v) == 0 {
			return "[]any"
		}

		type_ := t.getGoType(fieldName+"Item", v[0]) // TODO
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

	case nil:
		return "any"

	default:
		return "any"
	}
}

func (t *Transformer) toGoTypeName(fieldName string) string {
	goName := t.toGoFieldName(fieldName)
	if len(goName) > 0 {
		return strings.ToUpper(goName[:1]) + goName[1:]
	}
	return "Type"
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

func (t *Transformer) isTypeRecorded(name string) bool {
	for _, t := range t.types {
		if t.name == name {
			return true
		}
	}
	return false
}
