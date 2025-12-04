package singlevaluesubs

type Count[T any] struct {
	Broadcaster[T]
	count int
}

func (receiver *Count[T]) On_add(item T) {
	receiver.count++
	receiver.Broadcast(int64(receiver.count))
}

func (receiver *Count[T]) On_remove(item T) {
	receiver.count--
	receiver.Broadcast(int64(receiver.count))
}

func (receiver *Count[T]) On_update(oldItem T, newItem T) {
	receiver.count++
	receiver.Broadcast(int64(receiver.count))
}
