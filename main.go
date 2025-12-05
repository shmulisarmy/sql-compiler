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
	. "sql-compiler/rowType"
	"sql-compiler/state_full_byte_code"
	. "sql-compiler/tokenizer"
	option "sql-compiler/unwrap"
	. "sql-compiler/utils"
	"strconv"
)

type Table struct {
	Name    string
	Columns []rowType.ColInfo
	R_Table pubsub.R_Table
}

func (this *Table) hasCol(col_name string) bool {
	for i := range this.Columns {
		if this.Columns[i].Name == col_name {
			return true
		}
	}
	return false
}

func (this *Table) hasIndex(col_name string) bool {
	for i := range this.R_Table.Indexes {
		if this.R_Table.Indexes[i].Col_indexing_on == this.get_col_index(col_name) {
			return true
		}
	}
	return false
}

func (this *Table) Index_on(col_name string) *pubsub.Index {
	for i := range this.R_Table.Indexes {
		if this.R_Table.Indexes[i].Col_indexing_on == this.get_col_index(col_name) {
			return &this.R_Table.Indexes[i]
		}
	}
	display.DisplayStruct(this)
	this.R_Table.Indexes = append(this.R_Table.Indexes, pubsub.NewIndex(this.get_col_index(col_name), &this.R_Table))
	display.DisplayStruct(this)
	return &this.R_Table.Indexes[len(this.R_Table.Indexes)-1]
}

func (this *Table) insert(row rowType.RowType) {
	assert.AssertEq(len(row), len(this.Columns), fmt.Sprintf("rows in table %s must have %d columns and you passed a row that has %d columns", this.Name, len(this.Columns), len(row)))
	validate_col_types(this, &row)
	this.R_Table.Add(row)
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
		default:
			panic("unhandled")
		}
	}
}

func (this Table) get_index(col_name string) *pubsub.Index {
	for i := range this.R_Table.Indexes {
		if this.R_Table.Indexes[i].Col_indexing_on == this.get_col_index(col_name) {
			return &this.R_Table.Indexes[i]
		}
	}
	panic("col " + col_name + " not found in table " + this.Name)
}

func (this Table) get_col_index(col_name string) int {
	for i, col := range this.Columns {
		if col.Name == col_name {
			return i
		}
	}
	return -1
}

func get_Runtime_value_relative_location_and_type(select_ *ast.Select, col ast.Col) (byte_code.Runtime_value_relative_location, DataType) {
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
		table := tables.Get(select_.Table)
		index := table.get_col_index(col_name)
		if index != -1 {
			return byte_code.Runtime_value_relative_location{Amount_to_follow: 0, Col_index: index}, table.Columns[index].Type
		}
	}

Try_parent:
	if select_.Parent_select.IsNone() {
		panic("col " + col_name + " not found in select " + select_.Table)
	}
	location_info, type_ := get_Runtime_value_relative_location_and_type(select_.Parent_select.Unwrap(), col)
	return location_info.Add_one(), type_
}

func Recursively_set_selects_row_schema(select_ *ast.Select) RowSchema {
	for _, col := range select_.Selected_values {
		switch col_value := col.Value_to_select.(type) {
		case ast.Select:
			NestedSelectsRowSchema = append(NestedSelectsRowSchema, Recursively_set_selects_row_schema(&col_value))
			select_.Row_schema = append(select_.Row_schema, ColInfo{Name: col.Alias, Type: DataType(len(NestedSelectsRowSchema) - 1)})
		case ast.Plain_col_name:
			_, type_ := get_Runtime_value_relative_location_and_type(select_, col_value)
			schema_col_name := string(col_value)
			if col.Alias != "" {
				schema_col_name = col.Alias
			}
			select_.Row_schema = append(select_.Row_schema, ColInfo{Name: schema_col_name, Type: type_})
		case ast.Table_access:
			_, type_ := get_Runtime_value_relative_location_and_type(select_, col_value)
			schema_col_name := col_value.Col_name
			if col.Alias != "" {
				schema_col_name = col.Alias
			}
			select_.Row_schema = append(select_.Row_schema, ColInfo{Name: schema_col_name, Type: type_})
		//////////
		case int:
			select_.Row_schema = append(select_.Row_schema, ColInfo{Name: col.Alias, Type: Int})
		case string:
			select_.Row_schema = append(select_.Row_schema, ColInfo{Name: col.Alias, Type: String})
		case bool:
			select_.Row_schema = append(select_.Row_schema, ColInfo{Name: col.Alias, Type: Bool})
		default:
			panic("no other types supported")
		}
	}
	return select_.Row_schema
}
func get_Runtime_value_relative_location_if_Col(this *ast.Select, expr any) byte_code.Expression {
	if col, ok := expr.(ast.Col); ok {
		location_info, _ := get_Runtime_value_relative_location_and_type(this, col)
		return location_info
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
		switch col := col.Value_to_select.(type) {
		case ast.Select:
			// panic("not supported nested yet, coming soon...")
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, make_select_byte_code(&col))
		case ast.Plain_col_name:
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, get_Runtime_value_relative_location_if_Col(select_, col))
		case ast.Table_access:
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, get_Runtime_value_relative_location_if_Col(select_, col))
		case int, string, bool:
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, col)
		default:
			panic("unhandled")
		}
	}
	s.Col_and_value_to_index_by = tables.Get(select_.Table).choose_col_to_index(select_)
	return s
}

var tables = NewKeyValueArray[Table](30)

func init() {
	tables.Add("person", Table{
		Name:    "person",
		Columns: []ColInfo{{"name", String}, {"email", String}, {"age", Int}, {"state", String}, {"id", Int}},
		R_Table: pubsub.New_R_Table(),
	})
	tables.Add("todo", Table{
		Name:    "todo",
		Columns: []ColInfo{{"title", String}, {"description", String}, {"done", Bool}, {"person_id", Int}, {"is_public", Bool}},
		R_Table: pubsub.New_R_Table(),
	})
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
			panic(fmt.Sprintf("types %T and %T do not match", value1, value2))
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
			panic(fmt.Sprintf("types %T and %T do not match", value1, value2))
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
			panic(fmt.Sprintf("types %T and %T do not match", value1, value2))
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

func map_over(row_context state_full_byte_code.Row_context, selected_values_byte_code []byte_code.Expression, row_schema rowType.RowSchema) rowType.RowType {
	row := rowType.RowType{}
	for i, select_value_byte_code := range selected_values_byte_code { ///select_value_byte_code could just be a plain value
		switch select_value_byte_code := select_value_byte_code.(type) {
		case byte_code.Runtime_value_relative_location:
			row = append(row, row_context.Get_value(select_value_byte_code))
		case byte_code.Select:
			childs_row_context := state_full_byte_code.Row_context{Row: row_context.Row, Parent_context: option.Some(&row_context)}
			childs_row_schema := NestedSelectsRowSchema[row_schema[i].Type]
			row = append(row, select_byte_code_to_observable(select_value_byte_code, option.Some(&childs_row_context), childs_row_schema))
		default:
			row = append(row, select_value_byte_code)
		}
	}
	return row
}

func select_byte_code_to_observable(select_byte_code byte_code.Select, parent_context option.Option[*state_full_byte_code.Row_context], row_schema rowType.RowSchema) pubsub.ObservableI {
	var current_observable pubsub.ObservableI
	if select_byte_code.Col_and_value_to_index_by.Col != "" {
		//ints are cast to strings when placed and queried from indexes
		channel_value := select_byte_code.Col_and_value_to_index_by.Value
		switch channel_value := channel_value.(type) {
		case byte_code.Runtime_value_relative_location:
			tracked_channel_value := parent_context.Unwrap().Get_value(channel_value)
			current_observable = tables.Get(select_byte_code.Table_name).Index_on(select_byte_code.Col_and_value_to_index_by.Col).Get_or_create_channel_not_with_row(String_or_num_to_string(tracked_channel_value))
		case string:
			current_observable = tables.Get(select_byte_code.Table_name).Index_on(select_byte_code.Col_and_value_to_index_by.Col).Get_or_create_channel_not_with_row(channel_value)
		case int:
			int_str := strconv.Itoa(channel_value)
			current_observable = tables.Get(select_byte_code.Table_name).Index_on(select_byte_code.Col_and_value_to_index_by.Col).Get_or_create_channel_not_with_row(int_str)
		default:
			//bools are not supported for indexing indexes
			panic(fmt.Sprintf("%T %s", channel_value, channel_value))
		}
	} else {
		current_observable = &tables.Get(select_byte_code.Table_name).R_Table
	}

	current_observable = current_observable.Filter_on(func(row rowType.RowType) bool {
		return filter(state_full_byte_code.Row_context{Row: row, Parent_context: parent_context}, select_byte_code.Wheres_byte_code)
	}).Map_on(func(row rowType.RowType) rowType.RowType {
		return map_over(state_full_byte_code.Row_context{Row: row, Parent_context: parent_context}, select_byte_code.Selected_values_byte_code, row_schema)
	})

	current_observable.(*pubsub.Mapper).RowSchema = option.Some(row_schema)
	return current_observable

}

func main() {
	test_compilation()
}
func test_compilation() {
	tables.Get("person").Index_on("age")
	todos_table := tables.Get("todo")
	todos_table.Index_on("person_id")

	src := `SELECT person.email, person.name, person.id, (
		SELECT todo.title as epic_title, person.name as author, person.id FROM todo WHERE todo.person_id == person.id
		), (
		SELECT todo.title as epic_title FROM todo WHERE todo.is_public == true
		) as todo2 FROM person WHERE person.age > 3 `

	l := NewLexer(src)
	parser := parser{tokens: l.Tokenize()}
	for _, t := range parser.tokens {
		fmt.Printf("%-8s %q @%d\n", t.Type, t.Literal, t.Pos)
	}
	select_ := parser.parse_Select()
	select_.Recursively_link_children()
	Recursively_set_selects_row_schema(&select_)
	display.DisplayStruct(select_)
	select_byte_code := make_select_byte_code(&select_)
	display.DisplayStruct(select_byte_code)

	select_byte_code_to_observable(select_byte_code, option.None[*state_full_byte_code.Row_context](), select_.Row_schema).To_display(option.Some(select_.Row_schema))
	println(select_.Row_schema.To_string(0))

	todos_table.insert(rowType.RowType{"eat food", "make sure its clean", false, 1, false})
	todos_table.insert(rowType.RowType{"play music", "make sure its clean", false, 1, true})
	todos_table.insert(rowType.RowType{"clean", "make sure its clean", true, 1, false})
	todos_table.insert(rowType.RowType{"do art", "make sure its clean", false, 2, true})
	tables.Get("person").insert(rowType.RowType{"shmuli", "email@gmail.com", 25, "state", 1})
	tables.Get("person").insert(rowType.RowType{"the-doo-er", "email@gmail.com", 20, "state", 2})

	os.Exit(0)

}

func (table *Table) choose_col_to_index(select_ *ast.Select) byte_code.ColValuePair {
	type IndexSelectionInfo struct {
		channel_count int
		col_name      string
		value         any
	}

	best_index := IndexSelectionInfo{}
	for _, where := range select_.Wheres {
		var col string
		switch value1 := where.Value1.(type) {
		case ast.Plain_col_name:
			col = string(value1)
			if !table.hasCol(col) {
				continue
			}
			goto Try_to_index_col
		case ast.Table_access:
			if value1.Table_name != table.Name {
				continue
			}
			col = string(value1.Col_name)
			goto Try_to_index_col
		default:
			continue

		}
	Try_to_index_col:
		// if !table.col_is_primary_key(col) {
		// 	return byte_code.ColValuePair{
		// 		Col:   col,
		// 		Value: where.Value2,
		// 	}
		// }
		if !table.hasIndex(col) {
			continue
		}
		if where.Operator != EQ { //until we start using ordered maps then you can index on < and >
			continue
		}

		if _, is_of_type_bool := where.Value2.(bool); is_of_type_bool {
			continue
		}
		if table.get_index(col) != nil {
			best_index.channel_count = len(table.get_index(col).Channels)
			best_index.col_name = string(col)
			best_index.value = get_Runtime_value_relative_location_if_Col(select_, where.Value2)
		}
	}
	display.DisplayStruct(best_index)
	return byte_code.ColValuePair{
		Col:   best_index.col_name,
		Value: best_index.value,
	}
}
