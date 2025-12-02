package main

type Observable struct {
	Subscribers []Subscriber
}

func (this *Observable) add_sub(subscriber Subscriber) {
	this.Subscribers = append(this.Subscribers, subscriber)
}

func (this *Observable) publish_add(row RowType) {
	for _, subscriber := range this.Subscribers {

		subscriber.on_add(row)
	}
}

func (this *Observable) publish_remove(row RowType) {
	for _, subscriber := range this.Subscribers {
		subscriber.on_remove(row)
	}
}

func (this *Observable) publish_publish(old_row RowType, new_row RowType) {
	for _, subscriber := range this.Subscribers {
		subscriber.on_update(old_row, new_row)
	}
}

func link(observable ObservableI, subscriber Subscriber) {
	observable.add_sub(subscriber)
	subscriber.set_subscribed_to(observable)
}

type ObservableI interface {
	add_sub(subscriber Subscriber) //will get from Observable
	///
	pull(yield func(RowType) bool)
	publish_add(row RowType)
	publish_remove(row RowType)
	publish_publish(old_row RowType, new_row RowType)
	interface {
		filter_on(predicate func(RowType) bool) ObservableI
		map_on(transformer func(RowType) RowType) ObservableI
		to_display() *Printer
	}
}
type Subscriber interface {
	set_subscribed_to(observable ObservableI)
	///
	on_add(RowType)
	on_remove(RowType)
	on_update(old_row RowType, new_row RowType)
}
