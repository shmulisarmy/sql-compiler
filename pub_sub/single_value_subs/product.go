package singlevaluesubs

import "reflect"

type Product[T any] struct {
	Broadcaster[T]
	product        int64
	FieldToProduct string
}

func (receiver *Product[T]) On_add(item T) {
	receiver.product *= reflect.ValueOf(item).FieldByName(receiver.FieldToProduct).Int()
	receiver.Broadcast(int64(receiver.product))
}

func (receiver *Product[T]) On_remove(item T) {
	receiver.product /= reflect.ValueOf(item).FieldByName(receiver.FieldToProduct).Int()
	receiver.Broadcast(int64(receiver.product))
}

func (receiver *Product[T]) On_update(oldItem T, newItem T) {
	receiver.product /= reflect.ValueOf(oldItem).FieldByName(receiver.FieldToProduct).Int()
	receiver.product *= reflect.ValueOf(newItem).FieldByName(receiver.FieldToProduct).Int()
	receiver.Broadcast(int64(receiver.product))
}
