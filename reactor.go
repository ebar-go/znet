package znet

import (
	"context"
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal/poller"
	"log"
	"net"
)

// Reactor represents the epoll model for processing action connections.
type Reactor struct {
	poll     poller.Poller
	thread   *SubReactor
	engine   *Engine
	worker   pool.Worker
	callback *Callback
	
	packetLengthSize  int
	maxReadBufferSize int
}

// Run runs the Reactor with the given signal.
func (reactor *Reactor) Run(stopCh <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	// cancel context when the given signal is closed
	defer cancel()
	go func() {
		runtime.HandleCrash()
		// start thead polling task with active connection handler
		reactor.thread.Polling(ctx.Done(), reactor.onActive)
	}()

	reactor.run(stopCh)
}

// run receive active connection file descriptor and offer to thread
func (reactor *Reactor) run(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		default:
			// get the active connections
			active, err := reactor.poll.Wait()
			if err != nil {
				log.Println("unable to get active socket connection from epoll:", err)
				continue
			}

			// push the active connections to queue
			reactor.thread.Offer(active...)
		}
	}
}

// onConnect is called when the connection is established
func (reactor *Reactor) onConnect(conn net.Conn) {
	connection := NewConnection(conn, reactor.poll.SocketFD(conn))
	if err := reactor.poll.Add(connection.fd); err != nil {
		connection.Close()
		return
	}
	reactor.thread.RegisterConnection(connection)

	connection.AddBeforeCloseHook(
		reactor.callback.OnConnect,
		func(conn *Connection) {
			_ = reactor.poll.Remove(conn.fd)
		},
		reactor.thread.UnregisterConnection,
	)

	reactor.callback.OnDisconnect(connection)
}

// onActive is called when the connection is active
func (reactor *Reactor) onActive(fd int) {
	// receive an active connection
	conn := reactor.thread.GetConnection(fd)
	if conn == nil {
		return
	}

	// get bytes from pool, and release after processed
	bytes := pool.GetByte(reactor.maxReadBufferSize)
	// read message
	n, err := conn.ReadPacket(bytes, reactor.packetLengthSize)
	if err != nil {
		conn.Close()
		pool.PutByte(bytes)
		return
	}

	// process request
	reactor.worker.Schedule(func() {
		defer func() {
			runtime.HandleCrash()
			pool.PutByte(bytes)
		}()
		// prepare Context
		ctx := reactor.engine.NewContext(conn, bytes[:n])
		defer reactor.engine.ReleaseContext(ctx)

		reactor.engine.HandleContext(ctx)
	})
}

// ReactorOptions represents the options for the reactor
type ReactorOptions struct {
	// EpollBufferSize is the size of the active connections in every duration
	EpollBufferSize int

	// WorkerPollSize is the size of the worker pool
	WorkerPoolSize int

	// PacketLengthSize is the size of the packet length offset
	PacketLengthSize int

	// ThreadQueueCapacity is the cap of the thread queue
	ThreadQueueCapacity int

	MaxReadBufferSize int
}

func NewReactor(options ReactorOptions) (*Reactor, error) {
	poll, err := poller.NewPollerWithBuffer(options.EpollBufferSize)
	if err != nil {
		return nil, err
	}
	reactor := &Reactor{
		poll:              poll,
		engine:            NewEngine(),
		worker:            pool.NewGoroutinePool(options.WorkerPoolSize),
		packetLengthSize:  options.PacketLengthSize,
		maxReadBufferSize: options.MaxReadBufferSize,
		thread:            NewSubReactor(options.ThreadQueueCapacity),
	}

	return reactor, nil
}
