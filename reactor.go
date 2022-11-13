package znet

import (
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal/poller"
	"log"
)

// Reactor represents the epoll model for listen connections.
type Reactor struct {
	poll     poller.Poller // use to listen active connections
	sub      SubReactor    // manage connections
	callback *Callback     // manage connections events
}

// Run runs the Reactor with the given signal.
func (reactor *Reactor) Run(stopCh <-chan struct{}, onRequest ConnectionHandler) {
	subReactorSignal := make(chan struct{})
	defer close(subReactorSignal)
	go func() {
		defer runtime.HandleCrash()
		reactor.sub.Polling(subReactorSignal, reactor.wrapHandler(onRequest))
	}()

	pollerSignal := make(chan struct{})
	defer close(pollerSignal)
	go func() {
		defer runtime.HandleCrash()
		reactor.listenPoller(pollerSignal)
	}()

	runtime.WaitClose(stopCh)
}

func (reactor *Reactor) listenPoller(stopCh <-chan struct{}) {
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

func (reactor *Reactor) wrapHandler(handler ConnectionHandler) func(active int) {
	return func(active int) {
		conn := reactor.sub.GetConnection(active)
		if conn == nil {
			return
		}
		handler(conn)
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
		poll: poll,
		sub:  options.NewSubReactor(),
	}

	return reactor, nil
}
