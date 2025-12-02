package ast

import (
	. "sql-compiler/tokenizer"
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
