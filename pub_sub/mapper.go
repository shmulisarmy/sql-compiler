package pubsub

import (
	"sql-compiler/compiler/rowType"
	"sql-compiler/unwrap"
)

type Mapper struct {
	Observable
	transformer   func(rowType.RowType) rowType.RowType
	subscribed_to ObservableI
	RowSchema     unwrap.Option[rowType.RowSchema] //created when compiling the select, bases it off the tables (that were selecting from) schema and only places ones for the values that are actually being selected
}

func (this *Mapper) set_subscribed_to(observable ObservableI) {
	this.subscribed_to = observable
}

func (this *Mapper) Pull(yield func(rowType.RowType) bool) {
	for row := range this.subscribed_to.Pull {
		if !yield(this.transformer(row)) {
			return
		}
	}
}
func (this *Mapper) on_Add(row rowType.RowType) {
	this.Publish_Add(this.transformer(row))
}

func (this *Mapper) on_remove(row rowType.RowType) {
	this.Publish_remove(this.transformer(row))
}

func (this *Mapper) on_update(old_row rowType.RowType, new_row rowType.RowType) {
	this.Publish_Update(this.transformer(old_row), this.transformer(new_row))
}

func (this *Mapper) String() string {
	res := "["
	for row := range this.Pull {
		res += RowTypeToJson(&row, this.GetRowSchema()) + ","
	}
	return res + "]"
}
func (this *Mapper) GetRowSchema() rowType.RowSchema {
	return this.RowSchema.Unwrap()
}
