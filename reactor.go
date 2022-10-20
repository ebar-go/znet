package znet

import (
	"context"
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal/poller"
	"log"
)

// Reactor represents the epoll model for processing action connections.
type Reactor struct {
	poll              poller.Poller
	thread            *SubReactor
	engine            *Engine
	worker            pool.Worker
	packetLengthSize  int
	maxReadBufferSize int
}

// Run runs the Reactor with the given signal.
func (reactor *Reactor) Run(stopCh <-chan struct{}) {
	ctx := context.Background()

	threadCtx, threadCancel := context.WithCancel(ctx)
	// cancel context when the given signal is closed
	defer threadCancel()
	go func() {
		runtime.HandleCrash()
		// start thead polling task with active connection handler
		reactor.thread.Polling(threadCtx.Done(), reactor.handleActiveConnection)
	}()

	reactor.run(stopCh)
}

// run receive active connection file descriptor and offer to thread
func (reactor *Reactor) run(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		default:
			// get the active connections
			active, err := reactor.poll.Wait()
			if err != nil {
				log.Println("unable to get active socket connection from epoll:", err)
				continue
			}

			// push the active connections to queue
			reactor.thread.Offer(active...)
		}

	}
}

// handleActiveConnection handles active connection request
func (reactor *Reactor) handleActiveConnection(active int) {
	// receive an active connection
	conn := reactor.thread.GetConnection(active)
	if conn == nil {
		return
	}

	bytes := pool.GetByte(reactor.maxReadBufferSize)
	// read message
	n, err := conn.ReadPacket(bytes, reactor.packetLengthSize)
	if err != nil {
		conn.Close()
		pool.PutByte(bytes)
		return
	}

	// prepare Context
	ctx := reactor.engine.NewContext(conn, bytes[:n])

	// process request
	reactor.worker.Schedule(func() {
		reactor.engine.HandleContext(ctx)

		pool.PutByte(bytes)
	})
}

// ReactorOptions represents the options for the reactor
type ReactorOptions struct {
	// EpollBufferSize is the size of the active connections in every duration
	EpollBufferSize int

	// WorkerPollSize is the size of the worker pool
	WorkerPoolSize int

	// PacketLengthSize is the size of the packet length offset
	PacketLengthSize int

	// ThreadQueueCapacity is the cap of the thread queue
	ThreadQueueCapacity int

	MaxReadBufferSize int
}

func NewReactor(options ReactorOptions) (*Reactor, error) {
	poll, err := poller.NewPollerWithBuffer(options.EpollBufferSize)
	if err != nil {
		return nil, err
	}
	reactor := &Reactor{
		poll:              poll,
		engine:            NewEngine(),
		worker:            pool.NewGoroutinePool(options.WorkerPoolSize),
		packetLengthSize:  options.PacketLengthSize,
		maxReadBufferSize: options.MaxReadBufferSize,
		thread:            NewSubReactor(options.ThreadQueueCapacity),
	}

	return reactor, nil
}
