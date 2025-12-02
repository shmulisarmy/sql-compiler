package ast

import (
	. "sql-compiler/tokenizer"
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

type Select struct {
	Table           string
	Wheres          []Where
	Selected_values []any
	// compile time (post parsing stage) inserted
	Parent_select Option[*Select]
}

func (this *Select) Recursively_link_children() {
	print("sup")

	for i := range this.Selected_values {
		switch col := this.Selected_values[i].(type) {
		case Select:
			col.Parent_select = unwrap.Some(this)
			col.Recursively_link_children()
			this.Selected_values[i] = col
		case *Select:
			panic("unexpected pointer")
		}
	}
}
