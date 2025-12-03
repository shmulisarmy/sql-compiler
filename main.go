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
	"sql-compiler/state_full_byte_code"
	. "sql-compiler/tokenizer"
	option "sql-compiler/unwrap"
)

/////

type DataType int

const (
	String DataType = iota
	Int
	Bool
)

type ColInfo struct {
	Name string
	Type DataType
}

type Table struct {
	Name    string
	Columns []ColInfo
	r_Table pubsub.R_Table
}

func (this *Table) insert(row rowType.RowType) {
	assert.AssertEq(len(row), len(this.Columns), fmt.Sprintf("rows in table %s must have %d columns and you passed a row that has %d columns", this.Name, len(this.Columns), len(row)))
	validate_col_types(this, &row)
	this.r_Table.Add(row)
}

func validate_col_types(this *Table, row *rowType.RowType) {
	for i, col := range this.Columns {
		switch col.Type {
		case String:
			if _, ok := (*row)[i].(string); !ok {
				panic(fmt.Sprintf("col %s of table %s's type is string and you passed in a %T", col.Name, this.Name, (*row)[i]))
			}
		case Int:
			if _, ok := (*row)[i].(int); !ok {
				panic(fmt.Sprintf("col %s of table %s's type is int and you passed in a %T", col.Name, this.Name, (*row)[i]))
			}
		case Bool:
			if _, ok := (*row)[i].(bool); !ok {
				panic(fmt.Sprintf("col %s of table %s's type is bool and you passed in a %T", col.Name, this.Name, (*row)[i]))
			}
		}
	}
}

func (this Table) get_col_index(col_name string) int {
	println("trying to find " + col_name + " in " + this.Name)
	for i, col := range this.Columns {
		if col.Name == col_name {
			return i
		}
	}
	return -1
}

func get_Runtime_value_relative_location(select_ *ast.Select, col ast.Col) byte_code.Runtime_value_relative_location {
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
			return byte_code.Runtime_value_relative_location{Amount_to_follow: 0, Col_index: index}
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
		Columns: []ColInfo{{"name", String}, {"email", String}, {"age", Int}, {"state", String}, {"id", Int}},
		r_Table: pubsub.New_R_Table(),
	},
	"todo": {
		Name:    "todo",
		Columns: []ColInfo{{"title", String}, {"description", String}, {"done", Bool}, {"person_id", Int}, {"is_public", Bool}},
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
			return value1 == value2.(bool)
		default:
			panic("not implemented")
		}
	},
}

func filter(row_context state_full_byte_code.Row_context, wheres []byte_code.Where) bool {
	for _, where := range wheres {
		if !compare_methods[where.Compare_type](row_context.Track_value_if_is_relative_location(where.Value_1), row_context.Track_value_if_is_relative_location(where.Value_2)) {
			return false
		}
	}
	return true
}

func map_over(row_context state_full_byte_code.Row_context, selected_values_byte_code []byte_code.Expression) rowType.RowType {
	row := rowType.RowType{}
	for _, select_value_byte_code := range selected_values_byte_code { ///select_value_byte_code could just be a plain value
		switch select_value_byte_code := select_value_byte_code.(type) {
		case byte_code.Runtime_value_relative_location:
			row = append(row, row_context.Get_value(select_value_byte_code))
		case byte_code.Select:
			childs_row_context := state_full_byte_code.Row_context{Row: row_context.Row, Parent_context: option.Some(&row_context)}
			row = append(row, select_byte_code_to_observable(select_value_byte_code, option.Some(&childs_row_context)))
		default:
			row = append(row, select_value_byte_code)
		}
	}
	return row
}

func select_byte_code_to_observable(select_byte_code byte_code.Select, parent_context option.Option[*state_full_byte_code.Row_context]) pubsub.ObservableI {
	table := tables[select_byte_code.Table_name]
	return table.r_Table.Filter_on(func(row rowType.RowType) bool {
		return filter(state_full_byte_code.Row_context{Row: row, Parent_context: parent_context}, select_byte_code.Wheres_byte_code)
	}).Map_on(func(row rowType.RowType) rowType.RowType {
		return map_over(state_full_byte_code.Row_context{Row: row, Parent_context: parent_context}, select_byte_code.Selected_values_byte_code)
	})
}

func main() {
	src := `SELECT person.name, person.email, id, (
		SELECT todo.title, person.id FROM todo WHERE todo.is_public == true
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

	select_byte_code_to_observable(select_byte_code, option.None[*state_full_byte_code.Row_context]()).To_display()

	tables["todo"].insert(rowType.RowType{"clean", "make sure its clean", true, 1, false})
	tables["todo"].insert(rowType.RowType{"eat food", "make sure its clean", false, 1, false})
	tables["todo"].insert(rowType.RowType{"play music", "make sure its clean", false, 1, true})
	tables["todo"].insert(rowType.RowType{"do art", "make sure its clean", false, 2, true})
	tables["person"].insert(rowType.RowType{"shmuli", "email@gmail.com", 25, "state", 1})
	tables["person"].insert(rowType.RowType{"the-doo-er", "email@gmail.com", 20, "state", 2})

	os.Exit(0)

}
