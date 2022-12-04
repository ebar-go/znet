package znet

import (
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/codec"
	"log"
)

// Thread represents context manager
type Thread struct {
	options ThreadOptions
	codec   codec.Codec
	worker  pool.GoroutinePool
	engine  *Engine
}

// NewThread returns a new Thread instance
func NewThread(options ThreadOptions) *Thread {
	return &Thread{
		options: options,
		codec:   options.NewCodec(),
		worker:  options.NewWorkerPool(),
		engine:  NewEngine(),
	}
}

// Use registers middleware
func (thread *Thread) Use(handlers ...HandleFunc) {
	thread.engine.Use(handlers...)
}

// HandleRequest handle new request for connection
func (thread *Thread) HandleRequest(conn *Connection) {
	// read message from connection
	var (
		n      = 0
		bytes  = pool.GetByte(thread.options.MaxReadBufferSize)
		packet = codec.NewPacket(thread.codec)
	)

	err := runtime.Call(func() (lastErr error) {
		n, lastErr = conn.Read(bytes)
		return
	}, func() error {
		return packet.Unpack(bytes[:n])
	})

	if err != nil {
		log.Printf("[%s] read failed: %v\n", conn.ID(), err)
		// put back immediately when decode failed
		pool.PutByte(bytes)
		conn.Close()
		return
	}

	// compute
	thread.worker.Schedule(func() {
		defer runtime.HandleCrash()
		defer pool.PutByte(bytes)

		thread.engine.compute(conn, packet)
	})
}
