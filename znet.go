package znet

import (
	"context"
	"errors"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal"
	"log"
	"net"
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
	reactor *Reactor          // reactor model
	thread  *Thread           //
}

// New returns a new eng instance
func New(setters ...Option) Instance {
	options := completeOptions(setters...)

	return &Engine{
		options: options,
		reactor: options.NewReactorOrDie(),
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

	// compose work flow functions
	// decode request -> compute request -> encode response
	eng.thread.Use(eng.thread.decode(eng.router.handleError))
	eng.thread.Use(eng.options.Middlewares...)
	eng.thread.Use(eng.thread.compute(eng.router.handleRequest), eng.thread.encode(eng.router.handleError))

	// start listeners
	schemaCtx, schemeCancel := context.WithCancel(ctx)
	defer schemeCancel()
	if err := eng.startListenSchemas(schemaCtx); err != nil {
		return err
	}

	// start reactor
	reactorCtx, reactorCancel := context.WithCancel(ctx)
	defer reactorCancel()
	eng.startEventLoop(reactorCtx)

	runtime.WaitClose(stopCh, eng.shutdown)
	return nil
}

func (eng *Engine) startListenSchemas(ctx context.Context) error {
	// prepare servers
	for _, schema := range eng.schemas {
		// listen with context and connection register callback function
		if err := schema.Listen(ctx.Done(), func(conn net.Conn, protocol string) {
			// this callback will be invoked when the connection is established

			// create instance of Connection
			connection := NewConnection(conn, eng.reactor.poll.SocketFD(conn)).withProtocol(protocol)
			// initialize this new connection by reactor
			eng.reactor.initializeConnection(connection)
		}); err != nil {
			return err
		}

		log.Printf("start listener: %v\n", schema)
	}
	return nil
}

func (eng *Engine) startEventLoop(ctx context.Context) {
	go func() {
		defer runtime.HandleCrash()
		eng.reactor.Run(ctx.Done(), eng.thread.onRequest)
	}()
}

func (eng *Engine) shutdown() {
	log.Println("server shutdown complete")
}
