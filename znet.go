package znet

import (
	"context"
	"errors"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal"
	"log"
)

// Instance represents an reactor interface
type Instance interface {
	// Router return an router instance
	Router() *Router

	// Listen listens for different schema and address
	Listen(protocol string, addr string)

	// Run runs the instance with the given signal handler
	Run(stopCh <-chan struct{}) error
}

// Reactor represents im framework public access api.
type Reactor struct {
	options *Options
	schemas []internal.Schema
	router  *Router
	main    *MainReactor
}

// Listen register different protocols
func (reactor *Reactor) Listen(protocol string, addr string) {
	reactor.schemas = append(reactor.schemas, internal.NewSchema(protocol, addr))
}

// Router return instance of Router
func (reactor *Reactor) Router() *Router {
	return reactor.router
}

// Run starts the reactor
func (reactor *Reactor) Run(stopCh <-chan struct{}) error {
	ctx := context.Background()
	if len(reactor.schemas) == 0 {
		return errors.New("empty listen target")
	}

	// prepare servers
	schemaCtx, schemeCancel := context.WithCancel(ctx)
	// cancel schema context when reactor is stopped
	defer schemeCancel()
	for _, schema := range reactor.schemas {
		// listen with context and connection register callback function
		if err := schema.Listen(schemaCtx.Done(), reactor.main.onConnect); err != nil {
			return err
		}

		log.Printf("start listener: %v\n", schema)
	}

	// prepare reactor
	reactor.main.engine.Use(reactor.router.unpack)
	reactor.main.engine.Use(reactor.options.Middlewares...)
	reactor.main.engine.Use(reactor.router.handleRequest)

	reactorCtx, reactorCancel := context.WithCancel(ctx)
	// cancel reactor context when reactor is stopped
	defer reactorCancel()
	go func() {
		defer runtime.HandleCrash()
		reactor.main.Run(reactorCtx.Done())
	}()

	runtime.WaitClose(stopCh, reactor.shutdown)
	return nil
}

func (reactor *Reactor) shutdown() {
	log.Println("server shutdown complete")
}

// New returns a new reactor instance
func New(opts ...Option) Instance {
	options := defaultOptions()
	for _, setter := range opts {
		setter(options)
	}

	return &Reactor{
		options: options,
		main:    options.NewMainReactor(),
		router:  NewRouter(),
	}
}
