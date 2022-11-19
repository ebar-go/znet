package znet

import (
	"errors"
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/znet/codec"
	"github.com/ebar-go/znet/internal/acceptor"
	"time"
)

const (
	ContentTypeJson  = "json"
	ContentTypeProto = "protobuf"
)

// Options represents app options
type Options struct {
	// Debug enables debug logging
	Debug bool
	// OnOpen is a callback function that is called when the connection is established
	OnOpen ConnectionHandler

	// OnClose is a callback function that is called when the connection is closed
	OnClose ConnectionHandler

	// OnError is a callback function that is called when process error
	OnError func(ctx *Context, err error)

	// Middlewares is a lot of callback functions that are called when the connection send new message
	Middlewares []HandleFunc

	Reactor ReactorOptions

	Thread ThreadOptions

	Acceptor acceptor.Options
}

type ThreadOptions struct {
	// MaxReadBufferSize is the size of the max read buffer, default is 512
	MaxReadBufferSize int

	packetLengthSize int

	ContentType string

	WorkerPool *pool.Options
}

func (options ThreadOptions) NewWorkerPool() pool.WorkerPool {
	return pool.NewGoroutinePool(func(opts *pool.Options) {
		opts.Max = options.WorkerPool.Max
		opts.Idle = options.WorkerPool.Idle
		opts.Timeout = options.WorkerPool.Timeout
	})
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
	return reactor
}

func (options *Options) NewCallback() *Callback {
	return &Callback{
		openHandler:  options.OnOpen,
		closeHandler: options.OnClose,
		errorHandler: options.OnError,
	}
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

	if options.Thread.WorkerPool.Max <= 0 {
		return errors.New("Thread.WorkerPool.Max must be greater than zero")
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
		Debug:    false,
		OnOpen:   func(conn *Connection) {},
		OnClose:  func(conn *Connection) {},
		Reactor:  defaultReactorOptions(),
		Thread:   defaultThreadOptions(),
		Acceptor: acceptor.DefaultOptions(),
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
		packetLengthSize:  4,
		ContentType:       ContentTypeJson, // default is json
		WorkerPool: &pool.Options{
			Max:     10000,
			Idle:    100,
			Timeout: time.Minute,
		},
	}
}
