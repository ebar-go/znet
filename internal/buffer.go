package internal

// Buffer represents a buffer of any object
type Buffer[T any] struct {
	queue chan T
}

// Offer push the active fd to the queue, it will be deposed when buffer is full
func (buffer *Buffer[T]) Offer(items ...T) {
	for _, item := range items {
		// depose fd when queue is full
		select {
		case buffer.queue <- item:
		default:
			// must add default option,
			// otherwise it will:
			//		fatal error: all goroutines are asleep- deadlock!

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

// Empty returns true if the buffer is empty
func (buffer *Buffer[T]) Empty() bool {
	return buffer.Length() == 0
}

// Length returns the length of buffer
func (buffer *Buffer[T]) Length() int {
	return len(buffer.queue)
}

// NewBuffer returns a new buffer with the given size
func NewBuffer[T any](size int) *Buffer[T] {
	return &Buffer[T]{
		queue: make(chan T, size),
	}
}
