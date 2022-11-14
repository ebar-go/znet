package znet

import (
	"errors"
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/znet/codec"
)

const (
	ContentTypeJson  = "json"
	ContentTypeProto = "protobuf"
)

// Options represents app options
type Options struct {
	// Debug enables debug logging
	Debug bool
	// OnConnect is a callback function that is called when the connection is established
	OnConnect ConnectionHandler

	// OnDisconnect is a callback function that is called when the connection is closed
	OnDisconnect ConnectionHandler

	// Middlewares is a lot of callback functions that are called when the connection send new message
	Middlewares []HandleFunc

	Reactor ReactorOptions

	Thread ThreadOptions
}

type ThreadOptions struct {
	// WorkerPollSize is the size of the worker pool, default is 1000
	WorkerPoolSize int
	// MaxReadBufferSize is the size of the max read buffer, default is 512
	MaxReadBufferSize int

	packetLengthSize int

	ContentType string
}

func (options ThreadOptions) NewWorkerPool() pool.Worker {
	return pool.NewGoroutinePool(options.WorkerPoolSize)
}

func (options ThreadOptions) NewDecoder() codec.Decoder {
	return codec.NewDecoder(options.packetLengthSize)
}
func (options ThreadOptions) NewCodec() (cc codec.Codec) {
	cc = codec.NewJsonCodec()
	if options.ContentType == ContentTypeProto {
		cc = codec.NewProtoCodec()
	}
	return cc
}

// ReactorOptions represents the options for the reactor
type ReactorOptions struct {
	// EpollBufferSize is the size of the active connections in every duration,default is 100
	EpollBufferSize int

	// ThreadQueueCapacity is the cap of the thread queue, default is 100
	ThreadQueueCapacity int

	// SubReactorShardCount is the number of sub-reactor shards, default is 32
	// if the parameter is zero, the number of sub-reactor will be 1
	SubReactorShardCount int
}

func (options ReactorOptions) NewSubReactor() SubReactor {
	if options.SubReactorShardCount <= 1 {
		return NewSingleSubReactor(options.ThreadQueueCapacity)
	}

	return NewShardSubReactor(options.SubReactorShardCount, options.ThreadQueueCapacity)
}

func (options *Options) NewReactorOrDie() *Reactor {
	reactor, err := NewReactor(options.Reactor)
	if err != nil {
		panic(err)
	}
	reactor.callback = newCallback(options.OnConnect, options.OnDisconnect)
	return reactor
}

func (options *Options) NewThread() *Thread {
	return NewThread(options.Thread)
}

func (options *Options) NewRouter() *Router {
	return NewRouter()
}

// Validate validates the options parameter
func (options *Options) Validate() error {
	if options.Reactor.EpollBufferSize <= 0 {
		return errors.New("Reactor.EpollBufferSize must be greater than zero")
	}

	if options.Thread.WorkerPoolSize <= 0 {
		return errors.New("Thread.WorkerPoolSize must be greater than zero")
	}

	if options.Thread.MaxReadBufferSize <= 0 {
		return errors.New("Thread.MaxReadBufferSize must be greater than 0")
	}

	if options.Reactor.ThreadQueueCapacity <= 0 {
		return errors.New("Reactor.ThreadQueueCapacity must be greater than zero")
	}

	return nil
}

func completeOptions(setters ...Option) *Options {
	options := defaultOptions()
	for _, setter := range setters {
		setter(options)
	}
	return options
}

type Option func(options *Options)

func defaultOptions() *Options {
	return &Options{
		Debug:        false,
		OnConnect:    func(conn *Connection) {},
		OnDisconnect: func(conn *Connection) {},
		Reactor:      defaultReactorOptions(),
		Thread:       defaultThreadOptions(),
	}
}

func defaultReactorOptions() ReactorOptions {
	return ReactorOptions{
		EpollBufferSize:      256,
		ThreadQueueCapacity:  100,
		SubReactorShardCount: 32,
	}
}

func defaultThreadOptions() ThreadOptions {
	return ThreadOptions{
		MaxReadBufferSize: 512,
		WorkerPoolSize:    1000,
		packetLengthSize:  4,
		ContentType:       ContentTypeJson, // default is json
	}
}
