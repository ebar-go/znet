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
		options.OnOpen = func(conn *Connection) {
			log.Printf("[%s] connected", conn.ID())
		}
		options.OnClose = func(conn *Connection) {
			log.Printf("[%s] disconnected:%d", conn.ID(), time.Now().UnixMicro())
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	instance.ListenTCP(":8081")
	instance.ListenWebsocket(":8082")

	instance.Router().Route(1, func(ctx *Context) (any, error) {
		log.Printf("[%s] message: %s", ctx.Conn().ID(), string(ctx.Packet().Body))
		return map[string]any{"val": "bar"}, nil
	})
	err := instance.Run(ctx.Done())
	assert.Nil(t, err)

}

func TestClient(t *testing.T) {
	dialer := net.Dialer{KeepAlive: -1}
	conn, err := dialer.Dial("tcp", "localhost:8081")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	decoder := codec.NewDecoder(4)

	go func() {
		for {
			bytes, err := decoder.Decode(conn)
			if err != nil {
				log.Println("read error", time.Now().UnixMicro(), err)
				return
			}
			log.Println("receive response: ", string(bytes))
		}
	}()

	cc := codec.NewJsonCodec()

	packet := codec.NewPacket(cc)
	packet.Action = 1

	_ = packet.Marshal(map[string]interface{}{"foo": "bar"})
	p, _ := packet.Pack()

	log.Printf("packet length=%d, str=%s\n", len(p), string(p))

	go func() {
		for {
			n, err := decoder.Encode(conn, p)
			log.Println(n, err)
			time.Sleep(time.Second * 3)
		}
	}()
	select {}
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

	cc := codec.NewJsonCodec()

	packet := codec.NewPacket(cc)
	packet.Action = 1

	_ = packet.Marshal(map[string]interface{}{"foo": "bar"})
	bytes, _ := packet.Pack()
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
