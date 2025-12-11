// its up to the receiver to make sure that it stores the grabbed data in a grouped way, as now the way we send messages is with onAdd, .... with just rows (sending more data may be less efficient)
package pubsub

import (
	"sql-compiler/compiler/rowType"
	"sql-compiler/debugutil"
	"strconv"
)

type GroupBy struct {
	Observable
	subscribed_to            ObservableI
	index_of_col_to_group_by int //generated at compile time, to be used on RowSchema
	//another (more efficient (less S.O.L.I.D.), less solid way to do this is to just have this classes fields be public and have the event emmiter tree create the path directly instead of creating an entire observable for each separate group) way to do this is to have a map of the different groups and then when a row is added, removed or updated, we can just update the relevant group
	// different_groups map[string]Observable //this is so that the eventEmitterTree can create path based off how its grouped
}

func (this *GroupBy) Get_rows_group_value(row *rowType.RowType) string {
	// debugutil.Print(row, "row")
	// debugutil.Print(row, "is row, ")
	debugutil.Print(this.index_of_col_to_group_by, "this.index_of_col_to_group_by")
	rows_group_value := (*row)[this.index_of_col_to_group_by]
	switch rows_group_value := rows_group_value.(type) {
	case string:
		return rows_group_value
	case int:
		return strconv.Itoa(rows_group_value)
	default:
		panic("unexpected type")
	}
}
func (this *GroupBy) set_subscribed_to(observable ObservableI) {
	this.subscribed_to = observable
}

func (this *GroupBy) Pull(yield func(rowType.RowType) bool) {
	for row := range this.subscribed_to.Pull {
		if !yield(row) {
			return
		}
	}
}
func (this *GroupBy) on_Add(row rowType.RowType) {
	this.Publish_Add(row) //no point trying to combine the logic for the actual group because at the end of the day (according to the current architecture) its up to the receiver to make sure that it stores the grabbed data in a extra path'd way
}

func (this *GroupBy) on_remove(row rowType.RowType) {
	this.Publish_remove(row)
}

func (this *GroupBy) on_update(old_row rowType.RowType, new_row rowType.RowType) {
	this.Publish_Update(old_row, new_row)
}

func (this *GroupBy) String() string {
	res := "["
	for row := range this.Pull {
		res += RowTypeToJson(&row, this.GetRowSchema()) + ","
	}
	return res + "]"
}
func (this *GroupBy) GetRowSchema() rowType.RowSchema {
	return this.subscribed_to.GetRowSchema()
}
