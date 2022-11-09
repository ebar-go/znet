package znet

import (
	"context"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/codec"
	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	go func() {
		runtime.ShowMemoryUsage()
	}()
	instance := New(func(options *Options) {
		options.OnConnect = func(conn *Connection) {
			log.Printf("[%s] connected", conn.ID())
		}
		options.OnDisconnect = func(conn *Connection) {
			log.Printf("[%s] disconnected", conn.ID())
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	instance.ListenTCP(":8081")
	instance.ListenWebsocket(":8082")

	instance.Router().Route(1, func(ctx *Context) (any, error) {
		log.Printf("[%s] message: %s", ctx.Conn().ID(), string(ctx.Packet().Body()))
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

	bytes, err := codec.Factory().NewWithHeader(codec.Header{Operate: 1, Options: codec.OptionContentTypeJson}).Pack(map[string]any{"key": "foo"})
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
			default:

				n, err := conn.Write(bytes)
				log.Println(n, err)
				time.Sleep(time.Second)
			}

		}
	}()
	<-ctx.Done()
}

func BenchmarkClient(b *testing.B) {
	//runtime.SetLimit()
	opsRate := metrics.NewRegisteredTimer("ops", nil)

	ch := make(chan net.Conn, 200)
	n := 10000
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < 50; i++ {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					c, err := net.DialTimeout("tcp", ":8081", 10*time.Second)
					if err == nil {
						ch <- c
					}
				}
			}

		}(ctx)
	}
	connections := make([]net.Conn, 0, n)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:

			}
			time.Sleep(time.Second * 5)
			log.Println("connected:", len(connections))
		}
	}()
	for len(connections) < n {
		connections = append(connections, <-ch)
	}
	cancel()

	go func() {
		metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	}()

	bytes, err := codec.Factory().NewWithHeader(codec.Header{Operate: 1, Options: codec.OptionContentTypeJson}).Pack(map[string]any{"key": "foo"})
	if err != nil {
		return
	}

	b.ResetTimer()
	for i := 0; i < 100; i++ {
		go func() {
			for {
				n := rand.Intn(len(connections) - 1)
				c := connections[n]
				before := time.Now()
				if _, err := c.Write(bytes); err != nil {
					_ = c.Close()
					log.Println(err)
				} else {
					opsRate.Update(time.Now().Sub(before))
				}
			}

		}()
	}
	select {}
}
