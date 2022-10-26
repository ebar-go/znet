# znet
golang powerful network framework

## Features
- High-performance Event-Loop under multi-threads model
- Supporting multiple protocols: TCP,Websocket
- Supporting two event-notification mechanisms: epoll in Linux/Windows and kqueue in FreeBSD
- Supporting safe goroutines worker pool
- Supporting two contentType: JSON/Protobuf 

## Quick start
- install
```
go get -u github.com/ebar-go/znet
```

- go run server.go
```go
// server.go
package main

import (
	"github.com/ebar-go/ego/utils/runtime/signal"
	"github.com/ebar-go/znet"
)

const(
	OperateFoo = 1 // define a foo operate
)

func main() {
	// new instance
	instance := znet.New()

	// listen tcp and websocket
	instance.ListenTCP(":8081")
	instance.ListenWebsocket(":8082")
    
	// register a router for some operate
	instance.Router().Route(OperateFoo, func(ctx *znet.Context) (any, error) {
		// return response to the client
		return map[string]any{"val": "bar"}, nil
	})
	
	// run the instance, graceful stop by ctrl-c.
	instance.Run(signal.SetupSignalHandler())
}

```

- go run client.go

```go
// client.go
package main

import (
	"context"
	"github.com/ebar-go/znet/codec"
	"log"
	"net"
	"time"
)

func main() {
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
				return
			default:
				
			}
			packet := &codec.Packet{Header: codec.Header{Operate: OperateFoo, ContentType: codec.ContentTypeJSON}}
			bytes, err := codec.Default().Pack(packet, map[string]any{"key": "foo"})
			if err != nil {
				return
			}
			conn.Write(bytes)
			time.Sleep(time.Second)

		}
	}()
	<-ctx.Done()
}
```

## Protocol
- We design the protocol for solve TCP sticky packet problem
- ByteOrder: big endian
```
|-------------- header ------------- |-------- body --------|
|packetLength|operate|contentType|seq|-------- body --------|
|     4      |   2   |      2    | 2 |          n           |
```

## Benchmark
```
goos: linux
goarch: amd64
pkg: github.com/ebar-go/znet
cpu: Intel(R) Core(TM) i7-9700 CPU @ 3.00GHz

|-----------------------------------|
| connections  |  memory |    cpu   |
|-----------------------------------|
|     10000    |   50MB  |          |
|-----------------------------------|
```
