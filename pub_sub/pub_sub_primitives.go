package pubsub

import . "sql-compiler/rowType"

type Observable struct {
	Subscribers []Subscriber
}

func (this *Observable) Add_sub(subscriber Subscriber) {
	this.Subscribers = append(this.Subscribers, subscriber)
}

func (this *Observable) Publish_Add(row RowType) {
	for _, subscriber := range this.Subscribers {

		subscriber.on_Add(row)
	}
}

func (this *Observable) Publish_remove(row RowType) {
	for _, subscriber := range this.Subscribers {
		subscriber.on_remove(row)
	}
}

func (this *Observable) Publish_Publish(old_row RowType, new_row RowType) {
	for _, subscriber := range this.Subscribers {
		subscriber.on_update(old_row, new_row)
	}
}

func link(observable ObservableI, subscriber Subscriber) {
	observable.Add_sub(subscriber)
	subscriber.set_subscribed_to(observable)
}

type ObservableI interface {
	Add_sub(subscriber Subscriber) //will get from Observable
	///
	Pull(yield func(RowType) bool)
	Publish_Add(row RowType)
	Publish_remove(row RowType)
	Publish_Publish(old_row RowType, new_row RowType)
	interface {
		Filter_on(predicate func(RowType) bool) ObservableI
		Map_on(transformer func(RowType) RowType) ObservableI
		To_display() *Printer
	}
}
type Subscriber interface {
	set_subscribed_to(observable ObservableI)
	///
	on_Add(RowType)
	on_remove(RowType)
	on_update(old_row RowType, new_row RowType)
}
