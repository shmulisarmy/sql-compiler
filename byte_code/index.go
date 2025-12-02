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
