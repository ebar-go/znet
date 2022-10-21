package znet

import (
	"context"
	"github.com/ebar-go/znet/codec"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	instance := New(func(options *Options) {
		options.OnConnect = func(conn *Connection) {
			log.Printf("[%s] connected", conn.ID())
		}
		options.OnDisconnect = func(conn *Connection) {
			log.Printf("[%s] disconnected", conn.ID())
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	instance.ListenTCP(":8081")
	instance.ListenWebsocket(":8082")

	instance.Router().Route(1, func(ctx *Context) (any, error) {
		return map[string]any{"val": "bar"}, nil
	})
	err := instance.Run(ctx.Done())
	assert.Nil(t, err)

}

func TestClient(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	go func() {
		for {
			bytes := make([]byte, 512)
			n, err := conn.Read(bytes)
			if err != nil {
				return
			}
			log.Println("receive response: ", string(bytes[:n]))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
			default:
				packet := &codec.Packet{Operate: 1, ContentType: codec.ContentTypeJSON}
				bytes, err := codec.Default().Pack(packet, map[string]any{"key": "foo"})
				if err != nil {
					return
				}
				conn.Write(bytes)
				time.Sleep(time.Second)
			}

		}
	}()
	<-ctx.Done()
}
