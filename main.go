package main

import (
	"fmt"
	"os"
	"sql-compiler/assert"
	"sql-compiler/display"
	. "sql-compiler/tokenizer"
	option "sql-compiler/unwrap"
)

/////

type Table struct {
	Name    string
	Columns []string
	r_Table R_Table
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

func possibly_turn_from_string_into_bool(v any) any {
	return v
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

func (this *Select) recursively_link_children() {
	print("sup")

	for i := range this.Selected_values {
		switch col := this.Selected_values[i].(type) {
		case Select:
			col.Parent_select = option.Some(this)
			col.recursively_link_children()
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
	Table_name                string
	Wheres_byte_code          []Where_Byte_Code
	Selected_values_byte_code []Expression
}

func (this *Select) get_Runtime_value_relative_location_if_Col(expr any) Expression {
	if col, ok := expr.(Col); ok {
		return this.get_Runtime_value_relative_location(col)
	}
	return expr
}
func (this *Select) make_select_byte_code() Select_byte_code {
	assert.Assert(this.Table != "")
	s := Select_byte_code{
		Table_name: this.Table,
	}

	for _, where := range this.Wheres {
		s.Wheres_byte_code = append(s.Wheres_byte_code, Where_Byte_Code{
			Value_1:      this.get_Runtime_value_relative_location_if_Col(where.Value1),
			Compare_type: string(where.Operator),
			Value_2:      this.get_Runtime_value_relative_location_if_Col(where.Value2),
		})
	}

	for _, col := range this.Selected_values {
		switch col := col.(type) {
		case Select:
			// panic("not supported nested yet, coming soon...")
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, col.make_select_byte_code())
		case plain_col_name:
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, this.get_Runtime_value_relative_location_if_Col(col))
		case table_access:
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, this.get_Runtime_value_relative_location_if_Col(col))
		}
	}
	return s
}

var tables map[string]*Table = map[string]*Table{
	"person": &Table{
		Name:    "person",
		Columns: []string{"name", "email", "age", "state", "id"},
		r_Table: R_Table{
			Observable: Observable{
				Subscribers: []Subscriber{},
			},
			rows:       []RowType{},
			is_deleted: []bool{},
		},
	},
	"todo": &Table{
		Name:    "todo",
		Columns: []string{"title", "description", "done", "person_id"},
		r_Table: R_Table{
			Observable: Observable{
				Subscribers: []Subscriber{},
			},
			rows:       []RowType{},
			is_deleted: []bool{},
		},
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

var compare_methods = map[string]func(value1 any, value2 any) bool{

	"==": func(value1 any, value2 any) bool {
		switch value1 := value1.(type) {
		case string:
			return value1 == value2.(string)
		case int:
			return value1 == value2.(int)
		case bool:
			if value2 == "true" {
				value2 = true
			}
			if value2 == "false" {
				value2 = false
			}
			return value1 == value2.(bool)
		default:
			panic("not implemented")
		}
	},
	">": func(value1 any, value2 any) bool {
		switch value1 := value1.(type) {
		case string:
			return value1 > value2.(string)
		case int:
			return value1 > value2.(int)
		case bool:
			if value2 == "true" {
				value2 = true
			}
			if value2 == "false" {
				value2 = false
			}
			return value1 == value2.(bool)
		default:
			panic("not implemented")
		}
	},
	"<": func(value1 any, value2 any) bool {
		switch value1 := value1.(type) {
		case string:
			return value1 < value2.(string)
		case int:
			return value1 < value2.(int)
		case bool:
			if value2 == "true" {
				value2 = true
			}
			if value2 == "false" {
				value2 = false
			}
			return value1 == value2.(bool)
		default:
			panic("not implemented")
		}
	},
}

type Row_context struct {
	row            RowType
	parent_context option.Option[*Row_context]
}

func (this *Row_context) get_value(relative_location Runtime_value_relative_location) any {
	current := this
	for i := 0; i < relative_location.Amount_to_follow; i++ {
		current = current.parent_context.Expect(fmt.Sprintf("the fact that there is a problem with going up the stack on a relative_location.Amount_to_follow of %d is either a problem with linking in the parent context or a miscalculation on how far to go (a calculation made in func get_Runtime_value_relative_location as of 2025-12-02 in branch lsp)", relative_location.Amount_to_follow))
	}

	return current.row[relative_location.Col_index]
}

func (this *Row_context) track_value_if_is_relative_location(value any) any {
	if relative_location, ok := value.(Runtime_value_relative_location); ok {
		return this.get_value(relative_location)
	}
	return value
}

func filter(row_context Row_context, wheres []Where_Byte_Code) bool {
	for _, where := range wheres {
		if !compare_methods[where.Compare_type](row_context.track_value_if_is_relative_location(where.Value_1), row_context.track_value_if_is_relative_location(where.Value_2)) {
			return false
		}
	}
	return true
}

func map_over(row_context Row_context, selected_values_byte_code []Expression) RowType {
	row := RowType{}
	for _, select_value_byte_code := range selected_values_byte_code { ///select_value_byte_code could just be a plain value
		switch select_value_byte_code := select_value_byte_code.(type) {
		case Runtime_value_relative_location:
			row = append(row, row_context.get_value(select_value_byte_code))
		case Select_byte_code:
			childs_row_context := Row_context{row: row_context.row, parent_context: option.Some(&row_context)}
			row = append(row, select_byte_code_to_observable(select_value_byte_code, option.Some(&childs_row_context)))
		default:
			row = append(row, select_value_byte_code)
		}
	}
	return row
}

func select_byte_code_to_observable(select_byte_code Select_byte_code, parent_context option.Option[*Row_context]) ObservableI {
	table := tables[select_byte_code.Table_name]
	return table.r_Table.filter_on(func(row RowType) bool {
		return filter(Row_context{row: row, parent_context: parent_context}, select_byte_code.Wheres_byte_code)
	}).map_on(func(row RowType) RowType {
		return map_over(Row_context{row: row, parent_context: parent_context}, select_byte_code.Selected_values_byte_code)
	})
}

func main() {
	src := `SELECT person.name, person.email, id, (
		SELECT todo.title, person.id FROM todo WHERE todo.person_id == person.id  AND person.age > 22
		) FROM person WHERE person.age > 3 `

	l := NewLexer(src)
	parser := parser{tokens: l.Tokenize()}
	for _, t := range parser.tokens {
		fmt.Printf("%-8s %q @%d\n", t.Type, t.Literal, t.Pos)
	}
	select_ := parser.parse_Select()
	select_.recursively_link_children()
	display.DisplayStruct(select_)
	select_byte_code := select_.make_select_byte_code()
	display.DisplayStruct(select_byte_code)

	select_byte_code_to_observable(select_byte_code, option.None[*Row_context]()).to_display()

	tables["todo"].r_Table.add(RowType{"clean", "make sure its clean", true, 1})
	tables["todo"].r_Table.add(RowType{"eat food", "make sure its clean", false, 1})
	tables["todo"].r_Table.add(RowType{"play music", "make sure its clean", false, 1})
	tables["todo"].r_Table.add(RowType{"do art", "make sure its clean", false, 2})
	tables["person"].r_Table.add(RowType{"shmulik", "email@gmail.com", 25, "state", 1})
	tables["person"].r_Table.add(RowType{"baby chana", "email@gmail.com", 20, "state", 2})

	os.Exit(0)

}
