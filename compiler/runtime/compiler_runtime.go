package compiler_runtime

import (
	"fmt"
	"sql-compiler/compiler"
	"sql-compiler/compiler/parser"
	"sql-compiler/compiler/parser/tokenizer"
	"sql-compiler/compiler/rowType"
	"sql-compiler/compiler/state_full_byte_code"
	"sql-compiler/compiler/state_full_byte_code/byte_code"
	"sql-compiler/db_tables"
	"sql-compiler/display"
	pubsub "sql-compiler/pub_sub"
	option "sql-compiler/unwrap"
	"sql-compiler/utils"
	. "sql-compiler/utils"
	"strconv"
)

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
	">=": func(value1 any, value2 any) bool {
		switch value1 := value1.(type) {
		case string:
			return value1 >= value2.(string)
		case int:
			return value1 >= value2.(int)
		case bool:
			return value1 == value2.(bool)
		default:
			panic(fmt.Sprintf("types %T and %T do not match", value1, value2))
		}
	},
	"<=": func(value1 any, value2 any) bool {
		switch value1 := value1.(type) {
		case string:
			return value1 <= value2.(string)
		case int:
			return value1 <= value2.(int)
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
			childs_row_schema := rowType.NestedSelectsRowSchema[row_schema[i].Type]
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
			current_observable = db_tables.Tables.Get(select_byte_code.Table_name).Index_on(select_byte_code.Col_and_value_to_index_by.Col).Get_or_create_channel_not_with_row(String_or_num_to_string(tracked_channel_value))
		case string:
			current_observable = db_tables.Tables.Get(select_byte_code.Table_name).Index_on(select_byte_code.Col_and_value_to_index_by.Col).Get_or_create_channel_not_with_row(channel_value)
		case int:
			int_str := strconv.Itoa(channel_value)
			current_observable = db_tables.Tables.Get(select_byte_code.Table_name).Index_on(select_byte_code.Col_and_value_to_index_by.Col).Get_or_create_channel_not_with_row(int_str)
		default:
			//bools are not supported for indexing indexes
			panic(fmt.Sprintf("%T %s", channel_value, channel_value))
		}
	} else {
		current_observable = &db_tables.Tables.Get(select_byte_code.Table_name).R_Table
	}

	// Apply GROUP BY if specified (before Filter and Map)
	if select_byte_code.Group_by_col_index.IsSome() {
		current_observable = current_observable.GroupBy_on(select_byte_code.Group_by_col_index.Unwrap())
	}

	current_observable = current_observable.Filter_on(func(row rowType.RowType) bool {
		return filter(state_full_byte_code.Row_context{Row: row, Parent_context: parent_context}, select_byte_code.Wheres_byte_code)
	}).Map_on(func(row rowType.RowType) rowType.RowType {
		return map_over(state_full_byte_code.Row_context{Row: row, Parent_context: parent_context}, select_byte_code.Selected_values_byte_code, row_schema)
	})
	current_observable.(*pubsub.Mapper).RowSchema = option.Some(row_schema)
	if select_byte_code.Group_by_col_index.IsSome() {
		current_observable = current_observable.GroupBy_on(select_byte_code.Group_by_col_index.Unwrap())
	}
	return current_observable

}

func Query_to_observer(src string) pubsub.ObservableI {
	l := tokenizer.NewLexer(src)
	parser := parser.Parser{Tokens: l.Tokenize()}
	for _, t := range parser.Tokens {
		fmt.Printf("%-8s %q @%d\n", t.Type, t.Literal, t.Pos)
	}
	select_ := parser.Parse_Select()
	select_.Recursively_link_children()
	// display.DisplayStruct(select_)
	compiler.Recursively_set_selects_row_schema(&select_)
	select_byte_code := compiler.Make_select_byte_code(&select_)

	display.DisplayStruct(select_byte_code)

	obs := select_byte_code_to_observable(select_byte_code, option.None[*state_full_byte_code.Row_context](), select_.Row_schema)
	obs.To_display(option.Some(select_.Row_schema))
	fmt.Printf("type %s=%s\n", utils.Capitalize(select_.Table), select_.Row_schema.To_string(0))

	return obs
}
