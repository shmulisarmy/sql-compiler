package main

type Filter struct {
	Observable
	subscribed_to ObservableI
	predicate     func(RowType) bool
}

func (this *Filter) set_subscribed_to(observable ObservableI) {
	this.subscribed_to = observable
}

func (this *Filter) pull(yield func(RowType) bool) {
	this.subscribed_to.pull(func(row RowType) bool {
		if this.predicate(row) {
			yield(row)
		}
		return true
	})
}

func (this *Filter) on_add(row RowType) {
	if this.predicate(row) {
		this.publish_add(row)
	}
}

func (this *Filter) on_remove(row RowType) {
	if this.predicate(row) {
		this.publish_remove(row)
	}
}

func (this *Filter) on_update(old_row RowType, new_row RowType) {
	if this.predicate(new_row) {
		this.publish_publish(old_row, new_row)
	}
}
