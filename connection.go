package znet

import (
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	uuid "github.com/satori/go.uuid"
	"net"
	"sync"
)

type ConnectionHandler func(conn *Connection)

// Connection represents client connection
type Connection struct {
	// fd is the file descriptor
	fd int
	// uuid is the unique identifier
	uuid string
	// conn is the connection
	conn             net.Conn
	once             sync.Once
	beforeCloseHooks []func(connection *Connection)
	property         *Property
}

func (conn *Connection) Property() *Property {
	return conn.property
}

// UIID returns the uuid associated with the connection
func (conn *Connection) UUID() string { return conn.uuid }

// Push send message to the connection
func (conn *Connection) Push(p []byte) {
	conn.Write(p)
}

// Write writes message to the connection
func (conn *Connection) Write(p []byte) (int, error) {
	return conn.conn.Write(p)
}

// Read reads message from the connection
func (conn *Connection) Read(p []byte) (int, error) {
	return conn.conn.Read(p)
}

// Close closes the connection
func (conn *Connection) Close() {
	conn.once.Do(func() {
		for _, hook := range conn.beforeCloseHooks {
			hook(conn)
		}
		_ = conn.conn.Close()
	})
}

// AddBeforeCloseHook adds a hook to the connection before closed
func (conn *Connection) AddBeforeCloseHook(hooks ...func(conn *Connection)) {
	conn.beforeCloseHooks = append(conn.beforeCloseHooks, hooks...)
}

// readLine reads a line message from the connection
func (conn *Connection) readLine(buf []byte, packetLengthSize int) (n int, err error) {
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
func NewConnection(conn net.Conn, fd int) *Connection {
	return &Connection{conn: conn, fd: fd, uuid: uuid.NewV4().String(), property: &Property{properties: map[string]any{}}}
}

type Property struct {
	mu         sync.RWMutex // guards the properties
	properties map[string]any
}

func (p *Property) Set(key string, value any) {
	p.mu.Lock()
	p.properties[key] = value
	p.mu.Unlock()
}

func (p *Property) Get(key string) any {
	p.mu.RLock()
	property := p.properties[key]
	p.mu.RUnlock()
	return property
}

func (p *Property) GetString(key string) string {
	property := p.Get(key)
	if property == nil {
		return ""
	}
	return property.(string)
}
