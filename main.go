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

type Expression any

// interface {
// 	Expression__()
// }

type Where_Byte_Code struct {
	Value_1      Expression
	Compare_type string
	Value_2      Expression
}

type Select_byte_code struct {
	Wheres_byte_code          []Where_Byte_Code
	Selected_values_byte_code []Expression
}

func (this *Select) make_select_byte_code() Select_byte_code {
	s := Select_byte_code{}

	for _, where := range this.Wheres {
		s.Wheres_byte_code = append(s.Wheres_byte_code, Where_Byte_Code{
			Value_1:      this.get_Runtime_value_relative_location(where.Col),
			Compare_type: string(where.Operator),
			Value_2:      where.Value,
		})
	}

	for _, col := range this.Selected_values {
		switch col := col.(type) {
		case Select:
			// panic("not supported nested yet, coming soon...")
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, col.make_select_byte_code())
		case plain_col_name:
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, this.get_Runtime_value_relative_location(col))
		case table_access:
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, this.get_Runtime_value_relative_location(col))
		}
	}
	return s
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

func main1() {
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
func main() {
	src := `SELECT person.name, (SELECT person.state, (SELECT person.name FROM person WHERE todo.title > 18) FROM todo WHERE todo.done == true) FROM person WHERE person.age > 18`
	l := NewLexer(src)

	parser := parser{tokens: l.Tokenize()}
	for _, t := range parser.tokens {
		fmt.Printf("%-8s %q @%d\n", t.Type, t.Literal, t.Pos)
	}
	select_ := parser.parse_Select()
	select_.parent_children()
	display.Display(select_)
	select_byte_code := select_.make_select_byte_code()
	display.Display(select_byte_code)

	// Exit with success code
	os.Exit(0)

}
