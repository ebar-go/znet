package znet

import (
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/codec"
	"github.com/ebar-go/znet/internal"
)

// Thread represents context manager
type Thread struct {
	options ThreadOptions
	// handlerChains is a list of handlers
	handleChains []HandleFunc

	// contextProvider is a provider for context
	contextProvider internal.Provider[*Context]

	worker         pool.Worker
	codec          codec.Codec
	packetProvider internal.Provider[*codec.Packet]
	endian         binary.Endian
}

// Use registers middleware
func (e *Thread) Use(handler ...HandleFunc) {
	e.handleChains = append(e.handleChains, handler...)
}

// HandleRequest handle new request for connection
func (e *Thread) HandleRequest(conn *Connection) {
	// start schedule task
	// read request -> compute request -> send response
	e.worker.Schedule(func() {
		defer runtime.HandleCrash()

		// get bytes from pool, and release after processed
		bytes := pool.GetByte(e.options.MaxReadBufferSize)
		defer pool.PutByte(bytes)

		n, err := e.read(conn, bytes)
		if err != nil {
			conn.Close()
			return
		}
		// acquire context from provider
		ctx := e.contextProvider.Acquire()
		defer e.contextProvider.Release(ctx)

		// reset stateful properties
		ctx.reset(conn, bytes[:n])

		e.invokeContextHandler(ctx, 0)
	})
}

// ------------------------private methods------------------------

func (e *Thread) read(conn *Connection, bytes []byte) (n int, err error) {
	// read message
	if e.options.packetLengthSize == 0 {
		return conn.Read(bytes)
	}

	n, err = conn.Read(bytes[:e.options.packetLengthSize])
	if err != nil {
		return
	}
	packetLength := int(e.endian.Int32(bytes[:e.options.packetLengthSize]))
	if packetLength > len(bytes) {
		err = errors.New("packet exceeded")
		return
	}
	_, err = conn.Read(bytes[e.options.packetLengthSize:packetLength])
	n = packetLength
	return
}

func (e *Thread) decode(errorHandler func(ctx *Context, err error)) HandleFunc {
	return func(ctx *Context) {
		// new packet instance from pool, release it after finished
		packet := e.packetProvider.Acquire()
		defer e.packetProvider.Release(packet)
		packet.Reset()

		// unpack
		err := e.codec.Unpack(packet, ctx.msg)
		if err != nil {
			errorHandler(ctx, err)
			ctx.Abort()
			return
		}
		ctx.request = packet
		ctx.Next()
	}
}

func (e *Thread) encode(errorHandler func(*Context, error)) HandleFunc {
	return func(ctx *Context) {
		// choose not send response to client
		if ctx.response == nil {
			return
		}
		packet := ctx.Request()
		packet.Header.Seq++
		// pack response
		msg, err := e.codec.Pack(packet, ctx.response)
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

// NewThread returns a new Thread instance
func NewThread(options ThreadOptions) *Thread {
	engine := &Thread{
		options: options,
		worker:  pool.NewGoroutinePool(options.WorkerPoolSize),
		codec:   codec.Default(),
		packetProvider: internal.NewSyncPoolProvider[*codec.Packet](func() interface{} {
			return &codec.Packet{}
		}),
		endian: binary.BigEndian(),
	}

	engine.contextProvider = internal.NewSyncPoolProvider[*Context](func() interface{} {
		return &Context{thread: engine}
	})
	return engine
}
