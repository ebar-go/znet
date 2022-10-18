package internal

// Buffer represents a buffer of any object
type Buffer[T any] struct {
	queue chan T
}

// Offer push the active fd to the queue
func (buffer *Buffer[T]) Offer(items ...T) {
	for _, item := range items {
		// depose fd when queue is full
		select {
		case buffer.queue <- item:
		}
	}
}

// Polling poll with callback function
func (buffer *Buffer[T]) Polling(stopCh <-chan struct{}, handler func(item T)) {
	for {
		select {
		// stop when signal is closed
		case <-stopCh:
			return
		case active := <-buffer.queue:
			handler(active)
		default:
		}
	}
}

func NewBuffer[T any](size int) *Buffer[T] {
	return &Buffer[T]{
		queue: make(chan T, size),
	}
}
