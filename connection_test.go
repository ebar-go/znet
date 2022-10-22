package znet

import (
	"github.com/ebar-go/ego/utils/binary"
	"github.com/ebar-go/znet/internal"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"testing"
	"time"
)

func provideNetConn() net.Conn {
	stop := make(chan struct{})
	go internal.NewSchema(internal.TCP, ":8081").
		Listen(stop, func(conn net.Conn) {
			content := time.Now().String()
			buf := make([]byte, len(content)+4)
			binary.BigEndian().PutInt32(buf[:4], int32(len(content))+4)
			binary.BigEndian().PutString(buf[4:], content)
			conn.Write(buf)
		})

	time.Sleep(time.Second * 1)
	conn, err := net.Dial("tcp", ":8081")
	if err != nil {
		panic(err)
	}
	return conn
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
