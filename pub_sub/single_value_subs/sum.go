package singlevaluesubs

import "reflect"

type Sum[T any] struct {
	Broadcaster[T]
	sum        int64
	FieldToSum string
}

func (receiver *Sum[T]) On_add(item T) {
	receiver.sum += reflect.ValueOf(item).FieldByName(receiver.FieldToSum).Int()
	receiver.Broadcast(int64(receiver.sum))
}

func (receiver *Sum[T]) On_remove(item T) {
	receiver.sum -= reflect.ValueOf(item).FieldByName(receiver.FieldToSum).Int()
	receiver.Broadcast(int64(receiver.sum))
}

func (receiver *Sum[T]) On_update(oldItem T, newItem T) {
	receiver.sum -= reflect.ValueOf(oldItem).FieldByName(receiver.FieldToSum).Int()
	receiver.sum += reflect.ValueOf(newItem).FieldByName(receiver.FieldToSum).Int()
	receiver.Broadcast(int64(receiver.sum))
}
