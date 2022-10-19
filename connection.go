package znet

import (
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	"github.com/ebar-go/znet/internal"
	uuid "github.com/satori/go.uuid"
	"net"
	"sync"
)

// ConnectionHandler
type ConnectionHandler func(conn *Connection)

// Connection represents client connection
type Connection struct {
	// fd is the file descriptor
	fd int
	// uuid is the unique identifier
	uuid string
	// instance is the connection
	instance net.Conn
	// once make sure Close() is called only one times
	once sync.Once
	// beforeCloseHooks is a list of hooks that are called before the connection
	beforeCloseHooks []func(connection *Connection)
	// is a map of properties
	property *internal.Container[string, any]
}

// Property return properties container
func (conn *Connection) Property() *internal.Container[string, any] {
	return conn.property
}

// UIID returns the uuid associated with the connection
func (conn *Connection) ID() string { return conn.uuid }

// Push send message to the connection
func (conn *Connection) Push(p []byte) {
	conn.Write(p)
}

// Write writes message to the connection
func (conn *Connection) Write(p []byte) (int, error) {
	return conn.instance.Write(p)
}

// Read reads message from the connection
func (conn *Connection) Read(p []byte) (int, error) {
	return conn.instance.Read(p)
}

// Close closes the connection
func (conn *Connection) Close() {
	conn.once.Do(func() {
		for _, hook := range conn.beforeCloseHooks {
			hook(conn)
		}
		_ = conn.instance.Close()
	})
}

// AddBeforeCloseHook adds a hook to the connection before closed
func (conn *Connection) AddBeforeCloseHook(hooks ...func(conn *Connection)) {
	conn.beforeCloseHooks = append(conn.beforeCloseHooks, hooks...)
}

// ReadPacket reads a line message from the connection
func (conn *Connection) ReadPacket(buf []byte, packetLengthSize int) (n int, err error) {
	// if not set packetLengthSize, read buf directly
	if packetLengthSize == 0 {
		n, err = conn.Read(buf)
		return
	}

	// process tcp sticky package, read packet length first
	_, err = conn.Read(buf[:packetLengthSize])
	if err != nil {
		return
	}

	packetLength := int(binary.BigEndian().Int32(buf[:packetLengthSize]))
	if packetLength > len(buf) {
		err = errors.New("packet exceeded")
		return
	}
	_, err = conn.Read(buf[packetLengthSize:packetLength])
	n = packetLength
	return

}

// NewConnection returns a new Connection instance
func NewConnection(conn net.Conn, fd int) *Connection {
	return &Connection{
		instance: conn,
		fd:       fd,
		uuid:     uuid.NewV4().String(),
		property: internal.NewContainer[string, any](),
	}
}
