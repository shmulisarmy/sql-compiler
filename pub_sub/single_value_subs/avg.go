package singlevaluesubs

import (
	"reflect"
)

type Avg[T any] struct {
	Broadcaster[T]
	sum        int64
	totalCount int64
	FieldToSum string
}

func (receiver *Avg[T]) On_add(item T) {
	receiver.sum += reflect.ValueOf(item).FieldByName(receiver.FieldToSum).Int()
	receiver.totalCount++
	receiver.Broadcast(receiver.sum / receiver.totalCount)
}

func (receiver *Avg[T]) On_remove(item T) {
	receiver.sum -= reflect.ValueOf(item).FieldByName(receiver.FieldToSum).Int()
	receiver.totalCount--
	receiver.Broadcast(receiver.sum / receiver.totalCount)
}

func (receiver *Avg[T]) On_update(oldItem T, newItem T) {
	if receiver.totalCount < 1 {
		panic("your saying that you want to update an item, but there are no items in the collection")
	}
	receiver.sum -= reflect.ValueOf(oldItem).FieldByName(receiver.FieldToSum).Int()
	receiver.sum += reflect.ValueOf(newItem).FieldByName(receiver.FieldToSum).Int()
	receiver.Broadcast(receiver.sum / receiver.totalCount)
}
