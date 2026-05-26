package json2go

// ValueType the kind of json value represented by a [Value] node.
type ValueType int

const (
	NullValue ValueType = iota
	BoolValue
	StringValue
	NumberValue
	DecimalValue
	ObjectValue
	ArrayValue
)

// Value represents a json value in the AST.
type Value struct {
	Kind ValueType

	// only one of these is set depending on Kind
	Str    string
	Int    int64
	Float  float64
	Bool   bool
	Object []Field // ordered, preserves key order
	Array  []Value
}

type Field struct {
	K string
	V Value
}
