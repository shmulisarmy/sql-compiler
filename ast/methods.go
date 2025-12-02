package ast

import (
	"sql-compiler/assert"
	"sql-compiler/unwrap"
)

func (this *Select) Recursively_link_children() {
	print("sup")

	for i := range this.Selected_values {
		switch col := this.Selected_values[i].(type) {
		case Select:
			col.Parent_select = unwrap.Some(this)
			col.Recursively_link_children()
			this.Selected_values[i] = col
		case *Select:
			panic("unexpected pointer")
		}
	}
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
		case Plain_col_name:
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, this.get_Runtime_value_relative_location_if_Col(col))
		case Table_access:
			s.Selected_values_byte_code = append(s.Selected_values_byte_code, this.get_Runtime_value_relative_location_if_Col(col))
		}
	}
	return s
}

func (this *Select) get_Runtime_value_relative_location(col Col) Runtime_value_relative_location {
	var col_name string
	switch col := col.(type) {
	case Plain_col_name:
		col_name = string(col)
	case Table_access:
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
	return this.Parent_select.Unwrap().get_Runtime_value_relative_location(col).Add_one()
}
