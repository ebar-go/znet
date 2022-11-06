package znet

import (
	"context"
	"errors"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal"
	"log"
)

// Instance represents an eng interface
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

type Engine struct {
	options *Options          // options for the event loop
	schemas []internal.Schema // schema for acceptors
	router  *Router           // router for handlers
	main    *MainReactor
	thread  *Thread
}

// New returns a new eng instance
func New(opts ...Option) Instance {
	options := defaultOptions()
	for _, setter := range opts {
		setter(options)
	}

	return &Engine{
		options: options,
		main:    options.NewMainReactor(),
		router:  options.NewRouter(),
		thread:  options.NewThread(),
	}
}

// ListenTCP listens for tcp connections
func (eng *Engine) ListenTCP(addr string) {
	eng.schemas = append(eng.schemas, internal.NewSchema(internal.TCP, addr))
}

// ListenWebsocket listens for websocket connections
func (eng *Engine) ListenWebsocket(addr string) {
	eng.schemas = append(eng.schemas, internal.NewSchema(internal.WEBSOCKET, addr))
}

// Router return instance of Router
func (eng *Engine) Router() *Router {
	return eng.router
}

// Run starts the event-loop
func (eng *Engine) Run(stopCh <-chan struct{}) error {
	if err := eng.options.Validate(); err != nil {
		return err
	}
	ctx := context.Background()
	if len(eng.schemas) == 0 {
		return errors.New("please listen one protocol at least")
	}

	// prepare servers
	schemaCtx, schemeCancel := context.WithCancel(ctx)
	defer schemeCancel()
	if err := eng.runSchemas(schemaCtx); err != nil {
		return err
	}

	reactorCtx, reactorCancel := context.WithCancel(ctx)
	defer reactorCancel()
	eng.runReactor(reactorCtx)

	runtime.WaitClose(stopCh, eng.shutdown)
	return nil
}

func (eng *Engine) runSchemas(ctx context.Context) error {
	// prepare servers
	for _, schema := range eng.schemas {
		// listen with context and connection register callback function
		if err := schema.Listen(ctx.Done(), eng.main.onConnect); err != nil {
			return err
		}

		log.Printf("start listener: %v\n", schema)
	}
	return nil
}

func (eng *Engine) runReactor(ctx context.Context) {
	// decode request -> compute request -> encode response
	eng.thread.Use(eng.thread.decode(eng.router.handleError))
	eng.thread.Use(eng.options.Middlewares...)
	eng.thread.Use(eng.thread.compute(eng.router.handleRequest), eng.thread.encode(eng.router.handleError))

	go func() {
		defer runtime.HandleCrash()
		eng.main.Run(ctx.Done(), eng.thread.HandleRequest)
	}()
}

func (eng *Engine) shutdown() {
	log.Println("server shutdown complete")
}
