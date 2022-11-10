package znet

import (
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/codec"
	"github.com/ebar-go/znet/internal"
	"github.com/gobwas/ws/wsutil"
	"log"
)

// Thread represents context manager
type Thread struct {
	options ThreadOptions

	handleChains []HandleFunc // is a list of handlers

	contextProvider internal.Provider[*Context] // is a pool for Context
	worker          pool.Worker

	decoder codec.Decoder
}

// NewThread returns a new Thread instance
func NewThread(options ThreadOptions) *Thread {
	engine := &Thread{
		options: options,
		worker:  pool.NewGoroutinePool(options.WorkerPoolSize),
		decoder: codec.NewDecoder(options.packetLengthSize),
	}

	engine.contextProvider = internal.NewSyncPoolProvider[*Context](func() interface{} {
		return &Context{thread: engine}
	})
	return engine
}

// Use registers middleware
func (e *Thread) Use(handler ...HandleFunc) {
	e.handleChains = append(e.handleChains, handler...)
}

// HandleRequest handle new request for connection
func (e *Thread) HandleRequest(conn *Connection) {
	var (
		bytes         []byte
		msg           []byte
		err           error
		bytesFromPool bool
	)

	// read message from connection
	if conn.protocol == internal.WEBSOCKET {
		// read websocket request message
		msg, err = wsutil.ReadClientBinary(conn.instance)
	} else {
		var n int
		// get bytes from pool, and release after processed
		bytes = pool.GetByte(e.options.MaxReadBufferSize)
		bytesFromPool = true
		n, err = e.decoder.Decode(conn, bytes)
		if err == nil {
			msg = bytes[:n]
		}
	}

	// close the connection when read failed
	if err != nil {
		log.Printf("[%s] read: %v\n", conn.ID(), err)
		conn.Close()
		return
	}

	// start schedule task
	e.worker.Schedule(func() {
		defer runtime.HandleCrash()
		if bytesFromPool {
			defer pool.PutByte(bytes)
		}
		e.handleRequest(conn, msg)
	})

}

// ------------------------private methods------------------------

func (e *Thread) handleRequest(conn *Connection, msg []byte) {
	// close the connection when decode msg failed
	packet, err := codec.Factory().UnpackPacket(msg)
	if err != nil {
		log.Printf("[%s] decode: %v\n", conn.ID(), err)
		conn.Close()
		return
	}

	// compute
	// acquire context from provider
	ctx := e.contextProvider.Acquire()
	defer e.contextProvider.Release(ctx)

	ctx.reset(conn, packet)

	e.invokeContextHandler(ctx, 0)
}

func (e *Thread) encode(errorHandler func(*Context, error)) HandleFunc {
	return func(ctx *Context) {
		// pack response
		msg, err := ctx.Packet().Encode(ctx.response)
		if err != nil {
			errorHandler(ctx, err)
			return
		}

		ctx.Conn().Push(msg)
	}
}

// invokeContextHandler invoke context handler chain
func (e *Thread) invokeContextHandler(ctx *Context, index int8) {
	if int(index) > len(e.handleChains)-1 {
		return
	}
	e.handleChains[index](ctx)
}
