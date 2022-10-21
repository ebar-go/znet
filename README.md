# znet
golang powerful network framework

## Features
- High-performance Event-Loop under multi-threads model
- Supporting multiple protocols: TCP,Websocket
- Supporting two event-notification mechanisms: epoll in Linux/Windows and kqueue in FreeBSD
- Supporting safe goroutines worker pool

## Quick start
- install
```
go get -u github.com/ebar-go/znet
```

- run

```go
// server.go
package main

import (
	"github.com/ebar-go/ego/utils/runtime/signal"
	"github.com/ebar-go/znet/internal"
	"github.com/ebar-go/znet"
	"log"
)

func main() {
	instance := znet.New(znet.WithConnectCallback(func(conn *znet.Connection) {
		log.Printf("[%s] connected", conn.ID())
	}), znet.WithDisconnectCallback(func(conn *znet.Connection) {
		log.Printf("[%s] disconnected", conn.ID())
	}))


	instance.Listen(internal.TCP, ":8081")
	instance.Listen(internal.WEBSOCKET, ":8082")

	instance.Router().Route(1, func(ctx *znet.Context) (any, error) {
		return map[string]any{"val": "bar"}, nil
	})
	instance.Run(signal.SetupSignalHandler())
}
```

## Protocol

```
|-------------- header ------------- |-------- body --------|
|packetLength|operate|contentType|seq|-------- body --------|
|     4      |   2   |      2    | 2 |          n           |
```
