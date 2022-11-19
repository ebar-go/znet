package znet

import (
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

// Engine implements of Instance interface
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
	if len(eng.schemas) == 0 {
		return errors.New("please listen one protocol at least")
	}

	// compose work flow functions
	// decode request -> compute request -> encode response
	eng.thread.Use(eng.options.Middlewares...)
	eng.thread.Use(eng.router.handleRequest, eng.thread.encode(eng.router.triggerErrorEvent))

	// start listeners
	schemaSignal := make(chan struct{})
	defer close(schemaSignal)
	if err := eng.startListenSchemas(schemaSignal); err != nil {
		return err
	}

	// start reactor
	reactorSignal := make(chan struct{})
	defer close(reactorSignal)
	go func() {
		defer runtime.HandleCrash()
		eng.startEventLoop(reactorSignal)
	}()

	runtime.WaitClose(stopCh, eng.shutdown)
	return nil
}

// =====================private methods =================

func (eng *Engine) startListenSchemas(signal <-chan struct{}) error {
	// prepare servers
	for _, schema := range eng.schemas {
		// listen with context and connection register callback function
		if err := schema.Listen(signal, eng.reactor.initializeConnection); err != nil {
			return err
		}

		log.Printf("start listener: %v\n", schema)
	}
	return nil
}

func (eng *Engine) startEventLoop(signal <-chan struct{}) {
	eng.reactor.Run(signal, eng.thread.HandleRequest)
}

func (eng *Engine) shutdown() {
	log.Println("server shutdown complete")
}
