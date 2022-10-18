package znet

import (
	"github.com/ebar-go/znet/internal"
)

// Thread represents sub reactor
type Thread struct {
	buffer    *internal.Buffer[int]
	container *internal.Container[int, *Connection]
}

// RegisterConnection registers a new connection to the epoll listener
func (thread *Thread) RegisterConnection(conn *Connection) {
	thread.container.Set(conn.fd, conn)
}

// UnregisterConnection removes the connection from the epoll listener
func (thread *Thread) UnregisterConnection(conn *Connection) {
	thread.container.Del(conn.fd)
}

// GetConnection returns a connection by fd
func (thread *Thread) GetConnection(fd int) *Connection {
	conn, _ := thread.container.Get(fd)
	return conn
}

// Offer push the active connections fd to the queue
func (thread *Thread) Offer(fds ...int) {
	thread.buffer.Offer(fds...)

}

// Polling poll with callback function
func (thread *Thread) Polling(stopCh <-chan struct{}, handler func(active int)) {
	thread.buffer.Polling(stopCh, handler)
}

// NewThread creates a instance of a Thread
func NewThread(bufferSize int) *Thread {
	return &Thread{
		buffer:    internal.NewBuffer[int](bufferSize),
		container: internal.NewContainer[int, *Connection](),
	}
}
