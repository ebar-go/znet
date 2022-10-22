package znet

import (
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal"
)

// Thread represents context manager
type Thread struct {
	options ThreadOptions
	// handlerChains is a list of handlers
	handleChains []HandleFunc

	// contextProvider is a provider for context
	contextProvider internal.Provider[*Context]

	worker pool.Worker
}

// Use registers middleware
func (e *Thread) Use(handler ...HandleFunc) {
	e.handleChains = append(e.handleChains, handler...)
}

// HandleRequest handle new request for connection
func (e *Thread) HandleRequest(conn *Connection) {
	// get bytes from pool, and release after processed
	bytes := pool.GetByte(e.options.MaxReadBufferSize)
	// read message
	n, err := conn.ReadPacket(bytes, e.options.packetLengthSize)
	if err != nil {
		conn.Close()
		pool.PutByte(bytes)
		return
	}

	// process request
	e.worker.Schedule(func() {
		// avoid panic and release bytes
		defer func() {
			runtime.HandleCrash()
			pool.PutByte(bytes)
		}()
		// acquire context from provider
		ctx := e.contextProvider.Acquire()
		defer e.contextProvider.Release(ctx)

		// reset stateful properties
		ctx.reset(conn, bytes[:n])

		e.invokeContextHandler(ctx, 0)
	})
}

// ------------------------private methods------------------------

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
	}
	engine.contextProvider = internal.NewSyncPoolProvider[*Context](func() interface{} {
		return &Context{thread: engine}
	})
	return engine
}
