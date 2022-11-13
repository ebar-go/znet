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

	contextEngine *ContextEngine
	worker        pool.Worker

	decoder codec.Decoder
}

// NewThread returns a new Thread instance
func NewThread(options ThreadOptions) *Thread {
	thread := &Thread{
		options:       options,
		worker:        pool.NewGoroutinePool(options.WorkerPoolSize),
		decoder:       codec.NewDecoder(options.packetLengthSize),
		contextEngine: NewContextEngine(),
	}

	return thread
}

// Use registers middleware
func (thread *Thread) Use(handler ...HandleFunc) {
	thread.contextEngine.handleChains = append(thread.contextEngine.handleChains, handler...)
}

// HandleRequest handle new request for connection
func (thread *Thread) HandleRequest(conn *Connection) {
	var (
		msg      []byte
		err      error
		callback = func() {}
	)

	// read message from connection
	if conn.protocol == internal.WEBSOCKET {
		// read websocket request message
		msg, err = wsutil.ReadClientBinary(conn.instance)
	} else {
		var n int
		// get bytes from pool, and release after processed
		bytes := pool.GetByte(thread.options.MaxReadBufferSize)
		callback = func() {
			pool.PutByte(bytes)
		}
		n, err = thread.decoder.Decode(conn, bytes)
		if err == nil {
			msg = bytes[:n]
		}
	}

	// close the connection when read failed
	if err != nil {
		log.Printf("[%s] read: %v\n", conn.ID(), err)
		conn.Close()
		callback()
		return
	}

	// start schedule task
	thread.worker.Schedule(func() {
		defer runtime.HandleCrash()
		defer callback()
		// close the connection when decode msg failed
		packet, err := codec.Factory().UnpackPacket(msg)
		if err != nil {
			log.Printf("[%s] decode: %v\n", conn.ID(), err)
			conn.Close()
			return
		}

		// acquire context from provider
		ctx := thread.contextEngine.AcquireAndResetContext(conn, packet)
		defer thread.contextEngine.ReleaseContext(ctx)

		thread.contextEngine.invoke(ctx, 0)
	})

}

// ------------------------private methods------------------------

func (thread *Thread) encode(errorHandler func(*Context, error)) HandleFunc {
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
