package ast

import (
	. "sql-compiler/compiler/parser/tokenizer"
	"sql-compiler/compiler/rowType"
	"sql-compiler/unwrap"
	. "sql-compiler/unwrap"
)

type Col interface {
	__Col()
}

func (Plain_col_name) __Col() {}

type Plain_col_name string

type Table_access struct {
	Table_name string
	Col_name   string
}

func (Table_access) __Col() {}
func (Select) __Col()       {}

type Where struct {
	Value1   any
	Operator TokenType
	Value2   any
}

type Selected_value struct {
	Value_to_select any
	Alias           string
}

type Select struct {
	Table           string
	Wheres          []Where
	Selected_values []Selected_value
	///type info
	Row_schema rowType.RowSchema
	// compile time (post parsing stage) inserted
	Parent_select Option[*Select]
}

func (this *Select) Recursively_link_children() {
	for i := range this.Selected_values {
		switch col := this.Selected_values[i].Value_to_select.(type) {
		case Select:
			col.Parent_select = unwrap.Some(this)
			col.Recursively_link_children()
			this.Selected_values[i] = Selected_value{Value_to_select: col, Alias: this.Selected_values[i].Alias}
		case *Select:
			panic("unexpected pointer")
		}
	}
}
