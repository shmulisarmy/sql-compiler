package rowType

import (
	"strings"
)

type Actually_Col any
type RowType []Actually_Col
type DataType int
type RowSchema []ColInfo

var NestedSelectsRowSchema = []RowSchema{
	RowSchema{},
	RowSchema{},
	RowSchema{},
	RowSchema{},
	RowSchema{},
	RowSchema{},
} //DataType enum is also being resume to index into this array when its higher than all the regular enum values

const (
	String DataType = iota
	Int
	Bool
)

// any other enum values will be as a NestedSelect_index

func (r RowSchema) To_string(depth int) string {
	indent := strings.Repeat("  ", depth)
	childIndent := strings.Repeat("  ", depth+1)

	var b strings.Builder
	b.WriteString("{\n")

	for _, col := range r {
		b.WriteString(childIndent)
		b.WriteString(col.Name)
		b.WriteByte(' ')
		b.WriteString(col.Type.To_string(depth + 1))
		b.WriteByte('\n')
	}

	b.WriteString(indent)
	b.WriteByte('}')

	return b.String()
}

func (this DataType) To_string(depth int) string {
	switch this {
	case String:
		return "string"
	case Int:
		return "int"
	case Bool:
		return "bool"
	default:
		return NestedSelectsRowSchema[int(this)].To_string(depth + 1)
	}
}

type ColInfo struct {
	Name string
	Type DataType
}
