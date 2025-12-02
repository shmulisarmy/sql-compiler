package pubsub

import (
	"fmt"
	. "sql-compiler/rowType"
)

type Mapper struct {
	Observable
	transformer   func(RowType) RowType
	subscribed_to ObservableI
}

func (this *Mapper) set_subscribed_to(observable ObservableI) {
	this.subscribed_to = observable
}

func (this *Mapper) Pull(yield func(RowType) bool) {
	this.subscribed_to.Pull(func(row RowType) bool {
		yield(this.transformer(row))
		return true
	})
}
func (this *Mapper) on_Add(row RowType) {
	this.Publish_Add(this.transformer(row))
}

func (this *Mapper) on_remove(row RowType) {
	this.Publish_remove(this.transformer(row))
}

func (this *Mapper) on_update(old_row RowType, new_row RowType) {
	this.Publish_Publish(this.transformer(old_row), this.transformer(new_row))
}

func (this *Mapper) String() string {
	res := "["
	for row := range this.Pull {
		res += fmt.Sprintf("%v\n", row)
	}
	return res + "]"
}
