package znet

import (
	"context"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal/poller"
	"log"
	"net"
)

// MainReactor represents the epoll model for listen connections.
type MainReactor struct {
	options ReactorOptions
	// poll use to listen active connections
	poll poller.Poller

	// sub manage connections
	sub SubReactor

	// callback manage connections events
	callback *Callback
}

// Run runs the MainReactor with the given signal.
func (reactor *MainReactor) Run(stopCh <-chan struct{}, handler ConnectionHandler) {
	ctx, cancel := context.WithCancel(context.Background())
	// cancel context when the given signal is closed
	defer cancel()
	go func() {
		runtime.HandleCrash()
		// start sub reactor polling task with active connection handler
		reactor.sub.Polling(ctx.Done(), handler)
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

// NewMainReactor return a new main reactor instance
func NewMainReactor(options ReactorOptions) (*MainReactor, error) {
	poll, err := poller.NewPollerWithBuffer(options.EpollBufferSize)
	if err != nil {
		return nil, err
	}

	reactor := &MainReactor{
		options: options,
		poll:    poll,
	}

	// choose sub reactor implements by shard count
	if options.SubReactorShardCount <= 0 {
		reactor.sub = NewSingleSubReactor(options.ThreadQueueCapacity)
	} else {
		reactor.sub = NewShardSubReactor(options.SubReactorShardCount, options.ThreadQueueCapacity)
	}

	return reactor, nil
}
