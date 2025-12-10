package pubsub

import (
	"sql-compiler/compiler/rowType"
	. "sql-compiler/compiler/rowType"
)

type Filter struct {
	Observable
	subscribed_to ObservableI
	predicate     func(RowType) bool
}

func (this *Filter) set_subscribed_to(observable ObservableI) {
	this.subscribed_to = observable
}

func (this *Filter) Pull(yield func(RowType) bool) {
	for row := range this.subscribed_to.Pull {
		if this.predicate(row) {
			if !yield(row) {
				return
			}
		}
	}
}

func (this *Filter) on_Add(row RowType) {
	if this.predicate(row) {
		this.Publish_Add(row)
	}
}

func (this *Filter) on_remove(row RowType) {
	if this.predicate(row) {
		this.Publish_remove(row)
	}
}

func (this *Filter) on_update(old_row RowType, new_row RowType) {
	old_passed := this.predicate(old_row)
	new_passed := this.predicate(new_row)
	if old_passed && !new_passed {
		this.Publish_remove(old_row)
	} else if !old_passed && new_passed {
		this.Publish_Add(new_row)
	} else if old_passed && new_passed {
		this.Publish_Update(old_row, new_row)
	}
}

func (this *Filter) GetRowSchema() rowType.RowSchema {
	return this.subscribed_to.GetRowSchema()
}
