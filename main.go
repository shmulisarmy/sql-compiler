package main

import (
	"fmt"
	"os"
	"sql-compiler/assert"
	"sql-compiler/ast"
	"sql-compiler/byte_code"
	"sql-compiler/display"
	pubsub "sql-compiler/pub_sub"
	"sql-compiler/rowType"
	. "sql-compiler/tokenizer"
	option "sql-compiler/unwrap"
)

/////

type Table struct {
	Name    string
	Columns []string
	r_Table pubsub.R_Table
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

func (this Runtime_value_relative_location) Add_one() Runtime_value_relative_location {
	this.Amount_to_follow++
	return this
}

func possibly_turn_from_string_into_bool(v any) any {
	return v
}

func get_Runtime_value_relative_location(select_ *ast.Select, col ast.Col) Runtime_value_relative_location {
	var col_name string
	switch col := col.(type) {
	case ast.Plain_col_name:
		col_name = string(col)
	case ast.Table_access:
		col_name = col.Col_name
		if col.Table_name != select_.Table {
			goto Try_parent
		}
	case ast.Select:
		panic("not implemented")
	default:
		panic("unknown col type")
	}
	if col_name == "" {
		panic("col_name is empty")
	}
	{
		table := tables[select_.Table]
		index := table.get_col_index(col_name)
		if index != -1 {
			return Runtime_value_relative_location{Amount_to_follow: 0, Col_index: index}
		}
	}

Try_parent:
	if select_.Parent_select.IsNone() {
		panic("col " + col_name + " not found in select " + select_.Table)
	}
	return get_Runtime_value_relative_location(select_.Parent_select.Unwrap(), col).Add_one()
}

func get_Runtime_value_relative_location_if_Col(this *ast.Select, expr any) byte_code.Expression {
	if col, ok := expr.(ast.Col); ok {
		return get_Runtime_value_relative_location(this, col)
	}
	return expr
}
func make_select_byte_code(select_ *ast.Select) byte_code.Select {
	assert.Assert(select_.Table != "")
	s := byte_code.Select{
		Table_name: select_.Table,
	}

	for _, where := range select_.Wheres {
		s.Wheres_byte_code = append(s.Wheres_byte_code, byte_code.Where{
			Value_1:      get_Runtime_value_relative_location_if_Col(select_, where.Value1),
			Compare_type: string(where.Operator),
			Value_2:      get_Runtime_value_relative_location_if_Col(select_, where.Value2),
		})
	}

	for _, col := range select_.Selected_values {
		switch col := col.(type) {
		case ast.Select:
			// panic("not supported nested yet, coming soon...")
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, make_select_byte_code(&col))
		case ast.Plain_col_name:
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, get_Runtime_value_relative_location_if_Col(select_, col))
		case ast.Table_access:
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, get_Runtime_value_relative_location_if_Col(select_, col))
		}
	}
	return s
}

var tables = map[string]*Table{
	"person": {
		Name:    "person",
		Columns: []string{"name", "email", "age", "state", "id"},
		r_Table: pubsub.New_R_Table(),
	},
	"todo": {
		Name:    "todo",
		Columns: []string{"title", "description", "done", "person_id"},
		r_Table: pubsub.New_R_Table(),
	},
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
	row            rowType.RowType
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

func filter(row_context Row_context, wheres []byte_code.Where) bool {
	for _, where := range wheres {
		if !compare_methods[where.Compare_type](row_context.track_value_if_is_relative_location(where.Value_1), row_context.track_value_if_is_relative_location(where.Value_2)) {
			return false
		}
	}
	return true
}

func map_over(row_context Row_context, selected_values_byte_code []byte_code.Expression) rowType.RowType {
	row := rowType.RowType{}
	for _, select_value_byte_code := range selected_values_byte_code { ///select_value_byte_code could just be a plain value
		switch select_value_byte_code := select_value_byte_code.(type) {
		case Runtime_value_relative_location:
			row = append(row, row_context.get_value(select_value_byte_code))
		case byte_code.Select:
			childs_row_context := Row_context{row: row_context.row, parent_context: option.Some(&row_context)}
			row = append(row, select_byte_code_to_observable(select_value_byte_code, option.Some(&childs_row_context)))
		default:
			row = append(row, select_value_byte_code)
		}
	}
	return row
}

func select_byte_code_to_observable(select_byte_code byte_code.Select, parent_context option.Option[*Row_context]) pubsub.ObservableI {
	table := tables[select_byte_code.Table_name]
	return table.r_Table.Filter_on(func(row rowType.RowType) bool {
		return filter(Row_context{row: row, parent_context: parent_context}, select_byte_code.Wheres_byte_code)
	}).Map_on(func(row rowType.RowType) rowType.RowType {
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
	select_.Recursively_link_children()
	display.DisplayStruct(select_)
	select_byte_code := make_select_byte_code(&select_)
	display.DisplayStruct(select_byte_code)

	select_byte_code_to_observable(select_byte_code, option.None[*Row_context]()).To_display()

	tables["todo"].r_Table.Add(rowType.RowType{"clean", "make sure its clean", true, 1})
	tables["todo"].r_Table.Add(rowType.RowType{"eat food", "make sure its clean", false, 1})
	tables["todo"].r_Table.Add(rowType.RowType{"play music", "make sure its clean", false, 1})
	tables["todo"].r_Table.Add(rowType.RowType{"do art", "make sure its clean", false, 2})
	tables["person"].r_Table.Add(rowType.RowType{"shmulik", "email@gmail.com", 25, "state", 1})
	tables["person"].r_Table.Add(rowType.RowType{"baby chana", "email@gmail.com", 20, "state", 2})

	os.Exit(0)

}
