package compiler

import (
	"sql-compiler/compare"
	"sql-compiler/compiler/parser"
	"sql-compiler/compiler/parser/tokenizer"
	"sql-compiler/compiler/rowType"
	"testing"
)

func Test_query_to_row_schema(t *testing.T) {
	src := `select name, age from person `
	l := tokenizer.NewLexer(src)
	parser := parser.Parser{Tokens: l.Tokenize()}
	select_ := parser.Parse_Select()
	select_.Recursively_link_children()
	actual := Recursively_set_selects_row_schema(&select_)
	expected := rowType.RowSchema{
		rowType.ColInfo{Name: "name", Type: rowType.String},
		rowType.ColInfo{Name: "age", Type: rowType.Int},
	}
	output, err := compare.Compare(expected, actual, "")
	println(output)
	if err != nil {
		t.Error(err)
	}
}

func Test_query_to_row_schema_with_alias(t *testing.T) {
	src := `select name as persons_name, age from person `
	l := tokenizer.NewLexer(src)
	parser := parser.Parser{Tokens: l.Tokenize()}
	select_ := parser.Parse_Select()
	select_.Recursively_link_children()
	actual := Recursively_set_selects_row_schema(&select_)
	expected := rowType.RowSchema{
		rowType.ColInfo{Name: "persons_name", Type: rowType.String},
		rowType.ColInfo{Name: "age", Type: rowType.Int},
	}
	output, err := compare.Compare(expected, actual, "")
	println(output)
	if err != nil {
		t.Error(err)
	}
}
