# znet
golang powerful network framework


## Features
- High-performance Event-Loop under multi-threads model
- Supporting multiple protocols: TCP,Websocket
- Supporting reactor event-notification mechanisms: epoll in Linux/Windows and kqueue in FreeBSD
- Supporting safe goroutines worker pool
- Supporting two contentType: JSON/Protobuf 
- Supporting router service for different operate and handle functions



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
			packet := codec.Factory().NewWithHeader(codec.Header{Operate: OperateFoo, ContentType: codec.ContentTypeJSON})
			bytes, err := packet.Pack(map[string]any{"key": "foo"})
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


## Architecture
- Framework
  ![Framework](http://assets.processon.com/chart_image/62b3d00e637689074ac74fb3.png?1)
- Engine Start Sequence Diagram   
  ![Sequence Diagram](http://assets.processon.com/chart_image/6367a4755653bb5ba365c2ab.png?3)


## Protocol
- We design the protocol for solve TCP sticky packet problem
- ByteOrder: big endian
```
|-------------- header --------------|-------- body --------|
|packetLength| action |      seq     |-------- body --------|
|     4      |    2   |       2      |          n           |
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
