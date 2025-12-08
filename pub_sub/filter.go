package pubsub

import . "sql-compiler/rowType"

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
	if this.predicate(new_row) {
		this.Publish_Update(old_row, new_row)
	}
}
