package znet

import (
	"context"
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal/poller"
	"log"
	"net"
)

// MainReactor represents the epoll model for processing action connections.
type MainReactor struct {
	options ReactorOptions
	// poll use to listen active connections
	poll poller.Poller

	//
	sub      SubReactorInstance
	engine   *Engine
	worker   pool.Worker
	callback *Callback
}

// Run runs the MainReactor with the given signal.
func (reactor *MainReactor) Run(stopCh <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	// cancel context when the given signal is closed
	defer cancel()
	go func() {
		runtime.HandleCrash()
		// start thead polling task with active connection handler
		reactor.sub.Polling(ctx.Done(), reactor.onActive)
	}()

	reactor.run(stopCh)
}

// run receive active connection file descriptor and offer to thread
func (reactor *MainReactor) run(stopCh <-chan struct{}) {
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
			reactor.sub.Offer(active...)
		}
	}
}

// onConnect is called when the connection is established
func (reactor *MainReactor) onConnect(conn net.Conn) {
	connection := NewConnection(conn, reactor.poll.SocketFD(conn))
	if err := reactor.poll.Add(connection.fd); err != nil {
		connection.Close()
		return
	}
	reactor.sub.RegisterConnection(connection)

	connection.AddBeforeCloseHook(
		reactor.callback.OnDisconnect,
		func(conn *Connection) {
			_ = reactor.poll.Remove(conn.fd)
		},
		reactor.sub.UnregisterConnection,
	)

	reactor.callback.OnConnect(connection)
}

// onActive is called when the connection is active
func (reactor *MainReactor) onActive(fd int) {
	// receive an active connection
	conn := reactor.sub.GetConnection(fd)
	if conn == nil {
		return
	}

	// get bytes from pool, and release after processed
	bytes := pool.GetByte(reactor.options.MaxReadBufferSize)
	// read message
	n, err := conn.ReadPacket(bytes, reactor.options.packetLengthSize)
	if err != nil {
		conn.Close()
		pool.PutByte(bytes)
		return
	}

	// process request
	reactor.worker.Schedule(func() {
		// avoid panic and release bytes
		defer func() {
			runtime.HandleCrash()
			pool.PutByte(bytes)
		}()
		// handle request
		reactor.engine.HandleRequest(conn, bytes[:n])
	})
}

// NewMainReactor return a new main reactor instance
func NewMainReactor(options ReactorOptions) (*MainReactor, error) {
	poll, err := poller.NewPollerWithBuffer(options.EpollBufferSize)
	if err != nil {
		return nil, err
	}

	reactor := &MainReactor{
		options: options,
		poll:    poll,
		engine:  NewEngine(),
		worker:  pool.NewGoroutinePool(options.WorkerPoolSize),
	}

	if options.SubReactorShardCount <= 0 {
		reactor.sub = NewSubReactor(options.ThreadQueueCapacity)
	} else {
		reactor.sub = NewShardSubReactor(options.SubReactorShardCount, options.ThreadQueueCapacity)
	}

	return reactor, nil
}
