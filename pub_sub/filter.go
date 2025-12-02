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
	this.subscribed_to.Pull(func(row RowType) bool {
		if this.predicate(row) {
			yield(row)
		}
		return true
	})
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
		this.Publish_Publish(old_row, new_row)
	}
}
