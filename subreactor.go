package znet

import (
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal"
)

type SubReactor interface {
	RegisterConnection(conn *Connection)
	UnregisterConnection(conn *Connection)
	GetConnection(fd int) *Connection
	Offer(fds ...int)
	Polling(stopCh <-chan struct{}, callback func(int))
}

// SingleSubReactor represents sub reactor
type SingleSubReactor struct {
	// buffer manage active file descriptors
	buffer *internal.Buffer[int]

	// container manage all connections
	container *internal.Container[int, *Connection]
}

// RegisterConnection registers a new connection to the epoll listener
func (sub *SingleSubReactor) RegisterConnection(conn *Connection) {
	sub.container.Set(conn.fd, conn)
}

// UnregisterConnection removes the connection from the epoll listener
func (sub *SingleSubReactor) UnregisterConnection(conn *Connection) {
	sub.container.Del(conn.fd)
}

// GetConnection returns a connection by fd
func (sub *SingleSubReactor) GetConnection(fd int) *Connection {
	conn, _ := sub.container.Get(fd)
	return conn
}

// Offer push the active connections fd to the queue
func (sub *SingleSubReactor) Offer(fds ...int) {
	sub.buffer.Offer(fds...)
}

// Polling poll with callback function
func (sub *SingleSubReactor) Polling(stopCh <-chan struct{}, callback func(int)) {
	sub.buffer.Polling(stopCh, func(active int) {
		callback(active)
	})
}

// NewSingleSubReactor creates an instance of a SingleSubReactor
func NewSingleSubReactor(bufferSize int) *SingleSubReactor {
	return &SingleSubReactor{
		buffer:    internal.NewBuffer[int](bufferSize),
		container: internal.NewContainer[int, *Connection](),
	}
}

type ShardSubReactor struct {
	container internal.ShardContainer[*SingleSubReactor]
}

func (shard *ShardSubReactor) RegisterConnection(conn *Connection) {
	shard.container.GetShard(conn.fd).RegisterConnection(conn)
}

func (shard *ShardSubReactor) UnregisterConnection(conn *Connection) {
	shard.container.GetShard(conn.fd).UnregisterConnection(conn)
}

func (shard *ShardSubReactor) GetConnection(fd int) *Connection {
	return shard.container.GetShard(fd).GetConnection(fd)
}

func (shard *ShardSubReactor) Offer(fds ...int) {
	for _, fd := range fds {
		shard.container.GetShard(fd).Offer(fd)
	}
}

func (shard *ShardSubReactor) Polling(stopCh <-chan struct{}, callback func(int)) {
	shard.container.Iterator(func(sub *SingleSubReactor) {
		go func() {
			defer runtime.HandleCrash()
			sub.Polling(stopCh, callback)
		}()
	})
}

func NewShardSubReactor(shardCount, bufferSize int) *ShardSubReactor {
	return &ShardSubReactor{
		container: internal.NewShardContainer[*SingleSubReactor](shardCount, func() *SingleSubReactor {
			return NewSingleSubReactor(bufferSize)
		}),
	}
}
