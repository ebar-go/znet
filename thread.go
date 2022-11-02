package znet

import (
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/codec"
	"github.com/ebar-go/znet/internal"
	"github.com/gobwas/ws/wsutil"
)

// Thread represents context manager
type Thread struct {
	options ThreadOptions
	// handlerChains is a list of handlers
	handleChains []HandleFunc

	// contextProvider is a provider for context
	contextProvider internal.Provider[*Context]

	worker        pool.Worker
	codec         codec.Codec
	codecProvider internal.Provider[codec.Codec]
	endian        binary.Endian
}

// NewThread returns a new Thread instance
func NewThread(options ThreadOptions) *Thread {
	engine := &Thread{
		options: options,
		worker:  pool.NewGoroutinePool(options.WorkerPoolSize),
		codecProvider: internal.NewSyncPoolProvider[codec.Codec](func() interface{} {
			return codec.Default()
		}),
		endian: binary.BigEndian(),
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
	// start schedule task
	// read request -> compute request -> send response
	e.worker.Schedule(func() {
		defer runtime.HandleCrash()

		var (
			msg []byte
			err error
		)

		if conn.protocol == internal.TCP {
			var n int
			// get bytes from pool, and release after processed
			bytes := pool.GetByte(e.options.MaxReadBufferSize)
			defer pool.PutByte(bytes)

			n, err = e.read(conn, bytes)
			// reset stateful properties
			msg = bytes[:n]
		} else {
			msg, err = wsutil.ReadClientBinary(conn.instance)
		}

		if err != nil {
			conn.Close()
			return
		}

		// acquire context from provider
		ctx := e.contextProvider.Acquire()
		defer e.contextProvider.Release(ctx)

		codecInstance := e.codecProvider.Acquire()
		defer e.codecProvider.Release(codecInstance)
		ctx.reset(conn, msg)
		ctx.codec = codecInstance

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
	if packetLength < e.options.packetLengthSize || packetLength > len(bytes) {
		err = errors.New("packet exceeded")
		return
	}
	_, err = conn.Read(bytes[e.options.packetLengthSize:packetLength])
	n = packetLength
	return
}

func (e *Thread) decode(errorHandler func(ctx *Context, err error)) HandleFunc {
	return func(ctx *Context) {
		err := ctx.codec.Decode(ctx.request)
		if err != nil {
			errorHandler(ctx, err)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// compute process logic
func (e *Thread) compute(handler HandleFunc) HandleFunc {
	return handler
}

func (e *Thread) encode(errorHandler func(*Context, error)) HandleFunc {
	return func(ctx *Context) {
		// pack response
		msg, err := ctx.codec.Pack(ctx.response)
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
