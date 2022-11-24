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
	worker  pool.WorkerPool
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
	msg, err := thread.read(conn, true)
	if err != nil {
		log.Printf("[%s] read failed: %v\n", conn.ID(), err)
		// put back immediately when read failed
		pool.PutByte(msg)
		conn.Close()
		return
	}

	// decode packet from message
	packet, err := thread.decode(msg)
	if err != nil {
		log.Printf("[%s] decode failed: %v\n", conn.ID(), err)
		// put back immediately when decode failed
		pool.PutByte(msg)
		conn.Close()
		return
	}

	// compute
	thread.worker.Schedule(func() {
		defer runtime.HandleCrash()
		defer pool.PutByte(msg)

		thread.engine.compute(conn, packet)
	})
}

// ------------------------private methods------------------------
func (thread *Thread) read(conn *Connection, allocFromPool bool) (p []byte, err error) {
	var bytes []byte
	if allocFromPool {
		bytes = pool.GetByte(thread.options.MaxReadBufferSize)
	} else {
		bytes = make([]byte, thread.options.MaxReadBufferSize)
	}

	n, err := conn.Read(bytes)
	if err == nil {
		p = bytes[:n]
	}
	return
}

func (thread *Thread) decode(p []byte) (packet *codec.Packet, err error) {
	packet = codec.NewPacket(thread.codec)
	err = packet.Unpack(p)
	return
}
