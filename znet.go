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
	options *Options          // options for the event loop
	schemas []internal.Schema // schema for acceptors
	router  *Router           // router for handlers
	main    *MainReactor
	thread  *Thread
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
		return errors.New("please listen one protocol at least")
	}

	// prepare servers
	schemaCtx, schemeCancel := context.WithCancel(ctx)
	defer schemeCancel()
	if err := el.runSchemas(schemaCtx); err != nil {
		return err
	}

	// prepare handler func
	el.thread.Use(el.thread.decode(el.router.handleError))
	el.thread.Use(el.options.Middlewares...)
	el.thread.Use(el.thread.compute(el.router.handleRequest), el.thread.encode(el.router.handleError))

	reactorCtx, reactorCancel := context.WithCancel(ctx)
	defer reactorCancel()
	el.runReactor(reactorCtx)

	runtime.WaitClose(stopCh, el.shutdown)
	return nil
}

func (el *EventLoop) runSchemas(ctx context.Context) error {
	// prepare servers
	for _, schema := range el.schemas {
		// listen with context and connection register callback function
		if err := schema.Listen(ctx.Done(), el.main.onConnect); err != nil {
			return err
		}

		log.Printf("start listener: %v\n", schema)
	}
	return nil
}

func (el *EventLoop) runReactor(ctx context.Context) {
	go func() {
		defer runtime.HandleCrash()
		el.main.Run(ctx.Done(), el.thread.HandleRequest)
	}()
}

func (el *EventLoop) shutdown() {
	log.Println("server shutdown complete")
}
