package pubsub

import "sql-compiler/compiler/rowType"

// CustomSubscriber is a flexible subscriber that allows custom callback functions
type CustomSubscriber struct {
	Observable
	subscribed_to       ObservableI
	OnAddFunc           func(rowType.RowType)
	OnRemoveFunc        func(rowType.RowType)
	OnUpdateFunc        func(rowType.RowType, rowType.RowType)
	OnDeleteWhereEqFunc func(string, string)
}

// Compile-time interface checks
var _ Subscriber = (*CustomSubscriber)(nil)
var _ ObservableI = (*CustomSubscriber)(nil)

// NewCustomSubscriber creates a new CustomSubscriber with the provided callbacks
func NewCustomSubscriber(
	onAdd func(rowType.RowType),
	onRemove func(rowType.RowType),
	onUpdate func(rowType.RowType, rowType.RowType),
	onDeleteWhereEq func(string, string),
) *CustomSubscriber {
	return &CustomSubscriber{
		OnAddFunc:           onAdd,
		OnRemoveFunc:        onRemove,
		OnUpdateFunc:        onUpdate,
		OnDeleteWhereEqFunc: onDeleteWhereEq,
	}
}
func (this *CustomSubscriber) set_subscribed_to(observable ObservableI) {
	this.subscribed_to = observable
}

func (this *CustomSubscriber) Add_sub(subscriber Subscriber) {
	this.Subscribers = append(this.Subscribers, subscriber)
}

// on_Add is called when an item is added
func (receiver *CustomSubscriber) on_Add(item rowType.RowType) {
	if receiver.OnAddFunc != nil {
		receiver.OnAddFunc(item)
	} else {
		panic("OnAddFunc must be set")
	}
}

func (receiver *CustomSubscriber) DeleteWhereEq(key string, value string) {
	if receiver.OnDeleteWhereEqFunc != nil {
		receiver.OnDeleteWhereEqFunc(key, value)
	} else {
		panic("OnDeleteWhereEqFunc must be set")
	}
}

// on_remove is called when an item is removed
func (receiver *CustomSubscriber) on_remove(item rowType.RowType) {
	if receiver.OnRemoveFunc != nil {
		receiver.OnRemoveFunc(item)
	} else {
		panic("OnRemoveFunc must be set")
	}
}

// on_update is called when an item is updated
func (receiver *CustomSubscriber) on_update(oldItem, newItem rowType.RowType) {
	if receiver.OnUpdateFunc != nil {
		receiver.OnUpdateFunc(oldItem, newItem)
	} else {
		panic("OnUpdateFunc must be set")
	}
}

func (receiver *CustomSubscriber) Pull(yield func(rowType.RowType) bool) {
}
func (this *CustomSubscriber) GetRowSchema() rowType.RowSchema {
	return this.subscribed_to.GetRowSchema()
}
