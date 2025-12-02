package state_full_byte_code

import (
	"fmt"
	"sql-compiler/byte_code"
	"sql-compiler/rowType"
	option "sql-compiler/unwrap"
)

type Row_context struct {
	Row            rowType.RowType
	Parent_context option.Option[*Row_context]
}

func (this *Row_context) Get_value(relative_location byte_code.Runtime_value_relative_location) any {
	current := this
	for i := 0; i < relative_location.Amount_to_follow; i++ {
		current = current.Parent_context.Expect(fmt.Sprintf("the fact that there is a problem with going up the stack on a relative_location.Amount_to_follow of %d is either a problem with linking in the parent context or a miscalculation on how far to go (a calculation made in func get_Runtime_value_relative_location as of 2025-12-02 in branch lsp)", relative_location.Amount_to_follow))
	}

	return current.Row[relative_location.Col_index]
}

func (this *Row_context) Track_value_if_is_relative_location(value any) any {
	if relative_location, ok := value.(byte_code.Runtime_value_relative_location); ok {
		return this.Get_value(relative_location)
	}
	return value
}
