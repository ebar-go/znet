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

// Run starts the el
func (el *EventLoop) Run(stopCh <-chan struct{}) error {
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

	// prepare el
	el.main.engine.Use(el.router.unpack)
	el.main.engine.Use(el.options.Middlewares...)
	el.main.engine.Use(el.router.handleRequest)

	elCtx, elCancel := context.WithCancel(ctx)
	// cancel el context when el is stopped
	defer elCancel()
	go func() {
		defer runtime.HandleCrash()
		el.main.Run(elCtx.Done())
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
		router:  NewRouter(),
	}
}
