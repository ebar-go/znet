package acceptor

import (
	"net"
	"sync"
)

// Instance represents a server for accepting connections
type Instance interface {
	// Run runs the thread that will receive the connection
	Run(bind string) error

	// Shutdown shuts down the acceptor
	Shutdown()
}

type Acceptor struct {
	once    sync.Once
	done    chan struct{}
	handler func(conn net.Conn)
}

func (p *Acceptor) Signal() <-chan struct{} {
	return p.done
}

func (p *Acceptor) Done() {
	p.once.Do(func() {
		close(p.done)
	})
}

func NewAcceptor(handler func(conn net.Conn)) *Acceptor {
	return &Acceptor{
		once:    sync.Once{},
		done:    make(chan struct{}),
		handler: handler,
	}
}
