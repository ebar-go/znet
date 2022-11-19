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
	options  acceptor.Options
}

// String returns a string representation of the schema
func (schema Schema) String() string {
	return fmt.Sprintf("%s://%s", schema.Protocol, schema.Addr)
}

// Listen run acceptor with handler
func (schema Schema) Listen(stopCh <-chan struct{}, handler func(conn net.Conn)) error {
	var instance acceptor.Instance

	switch schema.Protocol {
	case TCP:
		instance = acceptor.NewTCPAcceptor(schema.options, handler)
	case WEBSOCKET:
		instance = acceptor.NewWSAcceptor(schema.options, handler)
	default:
		return fmt.Errorf("unsupported protocol: %v", schema.Protocol)
	}

	go func() {
		defer runtime.HandleCrash()
		runtime.WaitClose(stopCh, instance.Shutdown)
	}()
	return instance.Run(schema.Addr)
}

func NewSchema(protocol string, addr string, options acceptor.Options) Schema {
	return Schema{
		Protocol: protocol,
		Addr:     addr,
		options:  options,
	}
}
