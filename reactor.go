package znet

import (
	"context"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal/poller"
	"log"
)

// Reactor represents the epoll model for listen connections.
type Reactor struct {
	options  ReactorOptions
	poll     poller.Poller // use to listen active connections
	sub      SubReactor    // manage connections
	callback *Callback     // manage connections events
}

// Run runs the Reactor with the given signal.
func (reactor *Reactor) Run(stopCh <-chan struct{}, onRequest ConnectionHandler) {
	ctx, cancel := context.WithCancel(context.Background())
	// cancel context when the given signal is closed
	defer cancel()
	go func() {
		defer runtime.HandleCrash()
		// start sub reactor polling task with active connection handler
		reactor.sub.Polling(ctx.Done(), func(active int) {
			conn := reactor.sub.GetConnection(active)
			if conn == nil {
				return
			}
			onRequest(conn)
		})
	}()

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

			if len(active) == 0 {
				continue
			}

			// push the active connections to queue
			reactor.sub.Offer(active...)
		}
	}
}

func (reactor *Reactor) initializeConnection(connection *Connection) {
	if err := reactor.poll.Add(connection.fd); err != nil {
		connection.Close()
		return
	}

	reactor.callback.OnConnect(connection)

	reactor.sub.RegisterConnection(connection)

	// those callback functions will be invoked before connection.Close()
	connection.AddBeforeCloseHook(
		// trigger disconnect callback
		reactor.callback.OnDisconnect,
		// remove connection from epoll
		func(conn *Connection) {
			_ = reactor.poll.Remove(conn.fd)
		},
		// unregister connection from sub reactor
		reactor.sub.UnregisterConnection,
	)
}

// NewReactor return a new main reactor instance
func NewReactor(options ReactorOptions) (*Reactor, error) {
	poll, err := poller.NewPollerWithBuffer(options.EpollBufferSize)
	if err != nil {
		return nil, err
	}

	reactor := &Reactor{
		options: options,
		poll:    poll,
	}

	// choose sub reactor implements by shard count
	if options.SubReactorShardCount <= 1 {
		reactor.sub = NewSingleSubReactor(options.ThreadQueueCapacity)
	} else {
		reactor.sub = NewShardSubReactor(options.SubReactorShardCount, options.ThreadQueueCapacity)
	}

	return reactor, nil
}
