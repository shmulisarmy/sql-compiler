package byte_code

import "sql-compiler/unwrap"

type Expression any
type StringOrNumber any

type Where struct {
	Value_1      Expression
	Compare_type string
	Value_2      Expression
}

type ColValuePair struct {
	Col   string
	Value StringOrNumber
}

type Select struct {
	Table_name                string
	Col_and_value_to_index_by ColValuePair //could be empty, in which case we will take from the table directly
	Wheres_byte_code          []Where
	Selected_values_byte_code []Expression
	Group_by_col_index        unwrap.Option[int] // -1 if not set, otherwise the column index to group by
}

type Runtime_value_relative_location struct {
	Amount_to_follow int
	Col_index        int
}

func (this Runtime_value_relative_location) Add_one() Runtime_value_relative_location {
	this.Amount_to_follow++
	return this
}
