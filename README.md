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
	ActionFoo = 1 // define a foo operate
)

func main() {
	// new instance
	instance := znet.New()

	// listen tcp and websocket
	instance.ListenTCP(":8081")
	instance.ListenWebsocket(":8082")
    
	// register a router for some operate
	instance.Router().Route(ActionFoo, func(ctx *znet.Context) (any, error) {
		// return response to the client
		return map[string]any{"foo": "bar"}, nil
	})
	
	// run the instance, graceful stop by ctrl-c.
	instance.Run(signal.SetupSignalHandler())
}

```

## Architecture
- Framework
  ![Framework](http://assets.processon.com/chart_image/62b3d00e637689074ac74fb3.png?1)
- Engine Start Sequence Diagram   
  ![Sequence Diagram](http://assets.processon.com/chart_image/6367a4755653bb5ba365c2ab.png?3)


## Protocol
- TCP 
We design the protocol for solve TCP sticky packet problem
```
|-------------- header --------------|-------- body --------|
|packetLength| action |      seq     |-------- body --------|
|     4      |    2   |       2      |          n           |
```

- Websocket
websocket don't need the packet length
```
|-------------- header --------------|-------- body --------|
|        action       |      seq     |-------- body --------|
|           2         |       2      |          n           |
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
