package znet

import (
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/codec"
	"log"
)

// Thread represents context manager
type Thread struct {
	options       ThreadOptions
	codec         codec.Codec
	decoder       codec.Decoder
	worker        pool.WorkerPool
	contextEngine *ContextEngine
}

// NewThread returns a new Thread instance
func NewThread(options ThreadOptions) *Thread {
	thread := &Thread{
		options: options,
		codec:   options.NewCodec(),
		decoder: options.NewDecoder(),
		worker:  options.NewWorkerPool(),

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
		bytes    = pool.GetByte(thread.options.MaxReadBufferSize)
		packet   = thread.newPacket()
		callback = func() {
			pool.PutByte(bytes)
		}
	)

	err := func() (err error) {
		n, err := conn.Read(bytes)
		if err != nil {
			return
		}

		return packet.Unpack(bytes[:n])
	}()

	if err != nil {
		log.Printf("[%s] decode: %v\n", conn.ID(), err)
		conn.Close()
		callback()
		return
	}

	// start schedule task
	thread.worker.Schedule(func() {
		defer runtime.HandleCrash()
		defer callback()

		// acquire context from provider
		ctx := thread.contextEngine.AcquireAndResetContext(conn, packet)
		defer thread.contextEngine.ReleaseContext(ctx)

		thread.contextEngine.invoke(ctx, 0)
	})

}

// ------------------------private methods------------------------

func (thread *Thread) newPacket() *codec.Packet {
	return codec.NewPacket(thread.codec)
}
func (thread *Thread) encode(errorHandler func(*Context, error)) HandleFunc {
	return func(ctx *Context) {
		// pack response
		msg, err := ctx.packet.Pack()
		if err != nil {
			errorHandler(ctx, err)
			return
		}

		ctx.Conn().Write(msg)
	}
}
