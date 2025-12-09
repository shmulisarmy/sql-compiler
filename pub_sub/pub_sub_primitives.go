package pubsub

import (
	"sql-compiler/compiler/rowType"
	"sql-compiler/unwrap"
)

type Observable struct {
	Subscribers []Subscriber
}

func (this *Observable) Add_sub(subscriber Subscriber) {
	this.Subscribers = append(this.Subscribers, subscriber)
}

func (this *Observable) Publish_Add(row rowType.RowType) {
	for _, subscriber := range this.Subscribers {

		subscriber.on_Add(row)
	}
}

func (this *Observable) Publish_remove(row rowType.RowType) {
	for _, subscriber := range this.Subscribers {
		subscriber.on_remove(row)
	}
}

func (this *Observable) Publish_Update(old_row rowType.RowType, new_row rowType.RowType) {
	for _, subscriber := range this.Subscribers {
		subscriber.on_update(old_row, new_row)
	}
}

func Link(observable ObservableI, subscriber Subscriber) {
	observable.Add_sub(subscriber)
	subscriber.set_subscribed_to(observable)
}

type ObservableI interface {
	Add_sub(subscriber Subscriber) //will get from Observable
	///
	Pull(yield func(rowType.RowType) bool)
	Publish_Add(row rowType.RowType)
	Publish_remove(row rowType.RowType)
	Publish_Update(old_row rowType.RowType, new_row rowType.RowType)
	interface {
		Filter_on(predicate func(rowType.RowType) bool) ObservableI
		Map_on(transformer func(rowType.RowType) rowType.RowType) ObservableI
		To_display(unwrap.Option[rowType.RowSchema]) *Printer
	}
}
type Subscriber interface {
	set_subscribed_to(observable ObservableI)
	///
	on_Add(row rowType.RowType)
	on_remove(row rowType.RowType)
	on_update(old_row rowType.RowType, new_row rowType.RowType)
}
