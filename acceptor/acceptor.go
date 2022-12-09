package acceptor

import (
	"github.com/gobwas/ws"
	"net"
	"sync"
)

// Instance represents a server for accepting connections
type Instance interface {
	Schema() Schema
	// Listen runs the thread that will receive the connection
	Listen(onAccept func(conn net.Conn)) error

	// Shutdown shuts down the acceptor
	Shutdown()

	ReactorSupported() bool
}

type Acceptor struct {
	once   sync.Once
	done   chan struct{}
	schema Schema
}

func (acceptor *Acceptor) Schema() Schema {
	return acceptor.schema
}

func (acceptor *Acceptor) Shutdown() {
	acceptor.once.Do(func() {
		close(acceptor.done)
	})
}
func (acceptor *Acceptor) ReactorSupported() bool {
	return true
}

func NewAcceptor(schema Schema, options Options) Instance {
	acceptor := &Acceptor{
		schema: schema,
		once:   sync.Once{},
		done:   make(chan struct{}),
	}

	if schema.Protocol == TCP {
		return &TCPAcceptor{
			Acceptor: acceptor,
			options:  options,
		}
	} else if schema.Protocol == WEBSOCKET {
		return &WebsocketAcceptor{
			Acceptor: acceptor,
			options:  options,
			upgrade: ws.Upgrader{
				OnHeader: func(key, value []byte) (err error) {
					//log.Printf("non-websocket header: %q=%q", key, value)
					return
				},
			},
		}
	} else if schema.Protocol == QUIC {
		return &QUICAcceptor{
			Acceptor: acceptor,
			options:  options,
		}
	}

	return nil
}
