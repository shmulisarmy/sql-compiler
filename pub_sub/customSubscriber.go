package pubsub

import "sql-compiler/rowType"

// CustomSubscriber is a flexible subscriber that allows custom callback functions
type CustomSubscriber struct {
	Observable
	subscribed_to       ObservableI
	OnAddFunc           func(any)
	OnRemoveFunc        func(any)
	OnUpdateFunc        func(any, any)
	OnDeleteWhereEqFunc func(string, string)
}

// NewCustomSubscriber creates a new CustomSubscriber with the provided callbacks
func NewCustomSubscriber(
	onAdd func(any),
	onRemove func(any),
	onUpdate func(any, any),
) *CustomSubscriber {
	return &CustomSubscriber{
		OnAddFunc:    onAdd,
		OnRemoveFunc: onRemove,
		OnUpdateFunc: onUpdate,
	}
}
func (this *CustomSubscriber) set_subscribed_to(observable ObservableI) {
	this.subscribed_to = observable
}

func (this *CustomSubscriber) Add_sub(subscriber Subscriber) {
	this.Subscribers = append(this.Subscribers, subscriber)
}

// OnAdd is called when an item is added
func (receiver *CustomSubscriber) OnAdd(item any) {
	if receiver.OnAddFunc != nil {
		receiver.OnAddFunc(item)
	} else {
		panic("OnAddFunc must be set")
	}

}

func (receiver *CustomSubscriber) DeleteWhereEq(key string, value string) {
	if receiver.OnAddFunc != nil {
		receiver.OnDeleteWhereEqFunc(key, value)
	} else {
		panic("DeleteWhereEq must be set")
	}
}

// OnRemove is called when an item is removed
func (receiver *CustomSubscriber) OnRemove(item any) {
	if receiver.OnRemoveFunc != nil {
		receiver.OnRemoveFunc(item)
	} else {
		panic("OnRemoveFunc must be set")
	}
}

// OnUpdate is called when an item is updated
func (receiver *CustomSubscriber) OnUpdate(oldItem, newItem any) {
	if receiver.OnUpdateFunc != nil {
		receiver.OnUpdateFunc(oldItem, newItem)
	} else {
		panic("OnUpdateFunc must be set")
	}

}

func (receiver *CustomSubscriber) Pull(yield func(rowType.RowType) bool) {
	return
}
