package main

import (
	"fmt"
	"os"
	"sql-compiler/display"
	. "sql-compiler/tokenizer"
	option "sql-compiler/unwrap"
)

/////

type Table struct {
	Name    string
	Columns []string
}

func (this Table) get_col_index(col_name string) int {
	println("trying to find " + col_name + " in " + this.Name)
	for i, col := range this.Columns {
		if col == col_name {
			return i
		}
	}
	return -1
}

type Runtime_value_relative_location struct {
	Amount_to_follow int
	Col_index        int
}

func (this Runtime_value_relative_location) add_one() Runtime_value_relative_location {
	this.Amount_to_follow++
	return this
}

func (this *Select) get_Runtime_value_relative_location(col Col) Runtime_value_relative_location {
	var col_name string
	switch col := col.(type) {
	case plain_col_name:
		col_name = string(col)
	case table_access:
		col_name = col.Col_name
		if col.Table_name != this.Table {
			goto Try_parent
		}
	case Select:
		panic("not implemented")
	default:
		panic("unknown col type")
	}
	if col_name == "" {
		panic("col_name is empty")
	}
	{
		table := tables[this.Table]
		index := table.get_col_index(col_name)
		if index != -1 {
			return Runtime_value_relative_location{Amount_to_follow: 0, Col_index: index}
		}
	}

Try_parent:
	if this.Parent_select.IsNone() {
		panic("col " + col_name + " not found in select " + this.Table)
	}
	return this.Parent_select.Unwrap().get_Runtime_value_relative_location(col).add_one()
}

func (this *Select) parent_children() {
	print("sup")

	for i := range this.Selected_values {
		switch col := this.Selected_values[i].(type) {
		case Select:
			col.Parent_select = option.Some(this)
			col.parent_children()
			this.Selected_values[i] = col
		case *Select:
			panic("unexpected pointer")
		}
	}
}

func (this *Select) c() {

	wheres_byte_code := []Runtime_value_relative_location{}
	for _, where := range this.Wheres {
		wheres_byte_code = append(wheres_byte_code, this.get_Runtime_value_relative_location(where.Col))
	}
	selected_values_byte_code := []Runtime_value_relative_location{}

	for _, col := range this.Selected_values {
		switch col := col.(type) {
		case Select:
			col.c()
		case plain_col_name:
			selected_values_byte_code = append(selected_values_byte_code, this.get_Runtime_value_relative_location(col))
		case table_access:
			selected_values_byte_code = append(selected_values_byte_code, this.get_Runtime_value_relative_location(col))
		}
	}
	display.Display(wheres_byte_code)
	display.Display(selected_values_byte_code)
}

var tables map[string]Table = map[string]Table{
	"person": Table{
		Name:    "person",
		Columns: []string{"name", "email", "age", "state"},
	},
	"todo": Table{
		Name:    "todo",
		Columns: []string{"title", "description", "done", "person_id"},
	},
}

func main() {
	c := R_Table{}
	p := c.filter_on(func(row RowType) bool {
		return row[0] == "shmulik"
	}).map_on(func(rt RowType) RowType {
		return append(rt, "shmulik")
	}).to_display()
	c.add(RowType{"shmulik", "email@gmail.com", "25", "state"})
	c.add(RowType{"shmulik", "email@gmail.com", "25", "state"})
	p.run()
}
func main1() {
	src := `SELECT person.name, (SELECT person.state, (SELECT person.name FROM person WHERE todo.title > 18) FROM todo WHERE todo.done == true) FROM person WHERE person.age > 18`
	l := NewLexer(src)

	parser := parser{tokens: l.Tokenize()}
	for _, t := range parser.tokens {
		fmt.Printf("%-8s %q @%d\n", t.Type, t.Literal, t.Pos)
	}
	select_ := parser.parse_Select()
	select_.parent_children()
	display.Display(select_)
	select_.c()

	// Exit with success code
	os.Exit(0)

}
