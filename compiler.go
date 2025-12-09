package main

import (
	"sql-compiler/assert"
	"sql-compiler/ast"
	"sql-compiler/byte_code"
	"sql-compiler/db_tables"
	"sql-compiler/display"
	. "sql-compiler/parser/tokenizer"
	. "sql-compiler/rowType"
)

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
	table := db_tables.Tables.Get(select_.Table)
	s.Col_and_value_to_index_by = Choose_table_col_to_index(table, select_)
	return s
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
		table := db_tables.Tables.Get(select_.Table)
		index := table.Get_col_index(col_name)
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

func Choose_table_col_to_index(table *db_tables.Table, select_ *ast.Select) byte_code.ColValuePair {
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
			if !table.HasCol(col) {
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
		if !table.HasIndex(col) {
			continue
		}
		if where.Operator != EQ { //until we start using ordered maps then you can index on < and >
			continue
		}

		if _, is_of_type_bool := where.Value2.(bool); is_of_type_bool {
			continue
		}
		if table.Get_index(col) != nil {
			best_index.channel_count = len(table.Get_index(col).Channels)
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
