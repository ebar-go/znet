package znet

import (
	"context"
	"errors"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal"
	"log"
)

// Instance represents an el interface
type Instance interface {
	// Router return the router instance
	Router() *Router

	// ListenTCP listen tcp server
	ListenTCP(addr string)

	// ListenWebsocket listen websocket server
	ListenWebsocket(addr string)

	// Run runs the instance with the given signal handler
	Run(stopCh <-chan struct{}) error
}

// EventLoop represents im framework public access api.
type EventLoop struct {
	options *Options
	schemas []internal.Schema
	router  *Router
	main    *MainReactor
	thread  *Thread
}

// ListenTCP listens for tcp connections
func (el *EventLoop) ListenTCP(addr string) {
	el.schemas = append(el.schemas, internal.NewSchema(internal.TCP, addr))
}

// ListenWebsocket listens for websocket connections
func (el *EventLoop) ListenWebsocket(addr string) {
	el.schemas = append(el.schemas, internal.NewSchema(internal.WEBSOCKET, addr))
}

// Router return instance of Router
func (el *EventLoop) Router() *Router {
	return el.router
}

// Run starts the event-loop
func (el *EventLoop) Run(stopCh <-chan struct{}) error {
	if err := el.options.Validate(); err != nil {
		return err
	}
	ctx := context.Background()
	if len(el.schemas) == 0 {
		return errors.New("empty listen target")
	}

	// prepare servers
	schemaCtx, schemeCancel := context.WithCancel(ctx)
	// cancel schema context when el is stopped
	defer schemeCancel()
	for _, schema := range el.schemas {
		// listen with context and connection register callback function
		if err := schema.Listen(schemaCtx.Done(), el.main.onConnect); err != nil {
			return err
		}

		log.Printf("start listener: %v\n", schema)
	}

	// prepare handler func
	el.thread.Use(el.router.unpack)
	el.thread.Use(el.options.Middlewares...)
	el.thread.Use(el.router.handleRequest)

	reactorCtx, reactorCancel := context.WithCancel(ctx)
	// cancel reactor context when event-loop is stopped
	defer reactorCancel()
	go func() {
		defer runtime.HandleCrash()
		el.main.Run(reactorCtx.Done(), el.thread.HandleRequest)
	}()

	runtime.WaitClose(stopCh, el.shutdown)
	return nil
}

func (el *EventLoop) shutdown() {
	log.Println("server shutdown complete")
}

// New returns a new el instance
func New(opts ...Option) Instance {
	options := defaultOptions()
	for _, setter := range opts {
		setter(options)
	}

	return &EventLoop{
		options: options,
		main:    options.NewMainReactor(),
		router:  options.NewRouter(),
		thread:  options.NewThread(),
	}
}
