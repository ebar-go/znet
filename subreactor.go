package znet

import (
	"github.com/ebar-go/znet/internal"
)

// SubReactor represents sub reactor
type SubReactor struct {
	// buffer manage active file descriptors
	buffer *internal.Buffer[int]

	// container manage all connections
	container *internal.Container[int, *Connection]
}

// RegisterConnection registers a new connection to the epoll listener
func (sub *SubReactor) RegisterConnection(conn *Connection) {
	sub.container.Set(conn.fd, conn)
}

// UnregisterConnection removes the connection from the epoll listener
func (sub *SubReactor) UnregisterConnection(conn *Connection) {
	sub.container.Del(conn.fd)
}

// GetConnection returns a connection by fd
func (sub *SubReactor) GetConnection(fd int) *Connection {
	conn, _ := sub.container.Get(fd)
	return conn
}

// Offer push the active connections fd to the queue
func (sub *SubReactor) Offer(fds ...int) {
	sub.buffer.Offer(fds...)
}

// Polling poll with callback function
func (sub *SubReactor) Polling(stopCh <-chan struct{}, handler func(active int)) {
	sub.buffer.Polling(stopCh, handler)
}

// NewSubReactor creates a instance of a SubReactor
func NewSubReactor(bufferSize int) *SubReactor {
	return &SubReactor{
		buffer:    internal.NewBuffer[int](bufferSize),
		container: internal.NewContainer[int, *Connection](),
	}
}
