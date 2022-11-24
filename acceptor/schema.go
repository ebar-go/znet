package acceptor

import (
	"fmt"
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

func NewSchema(protocol string, addr string) Schema {
	return Schema{
		Protocol: protocol,
		Addr:     addr,
	}
}

func NewTCPSchema(addr string) Schema {
	return NewSchema(TCP, addr)
}
func NewWebSocketSchema(addr string) Schema {
	return NewSchema(WEBSOCKET, addr)
}
