package byte_code

type Expression any

type Where struct {
	Value_1      Expression
	Compare_type string
	Value_2      Expression
}

type Select struct {
	Table_name                string
	Wheres_byte_code          []Where
	Selected_values_byte_code []Expression
}

type Runtime_value_relative_location struct {
	Amount_to_follow int
	Col_index        int
}

func (this Runtime_value_relative_location) Add_one() Runtime_value_relative_location {
	this.Amount_to_follow++
	return this
}
