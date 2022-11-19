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
	worker        pool.WorkerPool
	contextEngine *ContextEngine
}

// NewThread returns a new Thread instance
func NewThread(options ThreadOptions) *Thread {
	thread := &Thread{
		options: options,
		codec:   options.NewCodec(),
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
	// read message from connection
	p, err := thread.read(conn)
	if err != nil {
		log.Printf("[%s] read failed: %v\n", conn.ID(), err)
		conn.Close()
		return
	}

	// decode packet from message
	packet, err := thread.decode(p)
	if err != nil {
		log.Printf("[%s] decode failed: %v\n", conn.ID(), err)
		conn.Close()
		return
	}

	// compute
	thread.worker.Schedule(func() {
		defer runtime.HandleCrash()
		defer pool.PutByte(p)

		// acquire context from provider
		ctx := thread.contextEngine.AcquireAndResetContext(conn, packet)
		defer thread.contextEngine.ReleaseContext(ctx)

		thread.contextEngine.invoke(ctx, 0)
	})

}

// ------------------------private methods------------------------
func (thread *Thread) read(conn *Connection) (p []byte, err error) {
	bytes := pool.GetByte(thread.options.MaxReadBufferSize)
	n, err := conn.Read(bytes)
	if err != nil {
		// put back immediately when read failed
		pool.PutByte(bytes)
		return
	}
	p = bytes[:n]
	return
}

func (thread *Thread) decode(p []byte) (packet *codec.Packet, err error) {
	packet = codec.NewPacket(thread.codec)
	err = packet.Unpack(p)
	if err != nil {
		// put back immediately when unpack failed
		pool.PutByte(p)
	}
	return
}
