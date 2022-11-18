package znet

import (
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"testing"
)

func provideNetConn() net.Conn {
	return nil
}
func TestNewConnection(t *testing.T) {
	connection := NewConnection(provideNetConn(), 1)
	assert.NotNil(t, connection)
	assert.NotEmpty(t, connection.ID())
}

func TestConnection_Close(t *testing.T) {
	connection := NewConnection(provideNetConn(), 1)
	connection.AddBeforeCloseHook(func(conn *Connection) {
		log.Println("called before closed")
	})
	connection.Close()

	// close again
	connection.Close()
}

func TestConnection_Property(t *testing.T) {
	connection := NewConnection(provideNetConn(), 1)
	connection.Property().Set("foo", "bar")

	item, exist := connection.Property().Get("foo")
	assert.True(t, exist)
	assert.Equal(t, "bar", item)
}

func TestConnection_Push(t *testing.T) {
	connection := NewConnection(provideNetConn(), 1)
	connection.Push([]byte("foo"))

	p := make([]byte, 512)
	n, err := connection.Read(p)
	assert.Nil(t, err)
	log.Println("receive:", string(p[:n]))
}
