package internal

import (
	"fmt"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal/acceptor"
	"net"
)

const (
	TCP       = "tcp"
	WEBSOCKET = "ws"
)

// Schema represents a protocol specification
type Schema struct {
	Protocol string
	Addr     string
}

// String returns a string representation of the schema
func (schema Schema) String() string {
	return fmt.Sprintf("%s://%s", schema.Protocol, schema.Addr)
}

// Listen run acceptor with handler
func (schema Schema) Listen(stopCh <-chan struct{}, handler func(conn net.Conn, protocol string)) error {
	var instance acceptor.Instance
	options := acceptor.DefaultOptions()
	switch schema.Protocol {
	case TCP:
		instance = acceptor.NewTCPAcceptor(options, func(conn net.Conn) {
			handler(conn, schema.Protocol)
		})
	case WEBSOCKET:
		instance = acceptor.NewWSAcceptor(options, func(conn net.Conn) {
			handler(conn, schema.Protocol)
		})
	default:
		return fmt.Errorf("unsupported protocol: %v", schema.Protocol)
	}

	go func() {
		defer runtime.HandleCrash()
		runtime.WaitClose(stopCh, instance.Shutdown)
	}()
	return instance.Run(schema.Addr)
}

func NewSchema(protocol string, addr string) Schema {
	return Schema{
		Protocol: protocol,
		Addr:     addr,
	}
}
