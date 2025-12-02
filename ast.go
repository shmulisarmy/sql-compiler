package main

import (
	. "sql-compiler/tokenizer"
	. "sql-compiler/unwrap"
)

type Col interface {
	__Col()
}

func (plain_col_name) __Col() {}

type plain_col_name string

type table_access struct {
	Table_name string
	Col_name   string
}

func (table_access) __Col() {}
func (Select) __Col()       {}

type where struct {
	Value1   any
	Operator TokenType
	Value2   any
}

type Select struct {
	Table           string
	Wheres          []where
	Selected_values []any
	// compile time inserted
	Parent_select Option[*Select]
}
