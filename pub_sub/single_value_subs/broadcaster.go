package singlevaluesubs

type Broadcaster[T any] struct {
	subscribers []func(int64)
}

func (receiver *Broadcaster[T]) Subscribe(subscriber func(int64)) *Broadcaster[T] {
	receiver.subscribers = append(receiver.subscribers, subscriber)
	return receiver
}

func (receiver *Broadcaster[T]) Broadcast(value int64) {
	for _, subscriber := range receiver.subscribers {
		subscriber(value)
	}
}
