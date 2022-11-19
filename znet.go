package znet

import (
	"errors"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal"
	"log"
)

// Instance implements of Instance interface
type Instance struct {
	options *Options          // options for the event loop
	schemas []internal.Schema // schema for acceptors
	router  *Router           // router for handlers
	reactor *Reactor          // reactor model
	thread  *Thread           //
}

// New returns a new instance instance
func New(setters ...Option) *Instance {
	options := completeOptions(setters...)

	return &Instance{
		options: options,
		reactor: options.NewReactorOrDie(),
		router:  options.NewRouter(),
		thread:  options.NewThread(),
	}
}

// ListenTCP listens for tcp connections
func (instance *Instance) ListenTCP(addr string) {
	instance.schemas = append(instance.schemas, internal.NewSchema(internal.TCP, addr, instance.options.Acceptor))
}

// ListenWebsocket listens for websocket connections
func (instance *Instance) ListenWebsocket(addr string) {
	instance.schemas = append(instance.schemas, internal.NewSchema(internal.WEBSOCKET, addr, instance.options.Acceptor))
}

// Router return instance of Router
func (instance *Instance) Router() *Router {
	return instance.router
}

// Run starts the event-loop
func (instance *Instance) Run(stopCh <-chan struct{}) error {
	if err := instance.options.Validate(); err != nil {
		return err
	}
	if len(instance.schemas) == 0 {
		return errors.New("there are no listeners available")
	}

	instance.thread.Use(instance.options.Middlewares...)
	instance.thread.Use(instance.router.handleRequest)

	// start listeners
	listenerSignal := make(chan struct{})
	defer close(listenerSignal)
	if err := instance.startListenSchemas(listenerSignal); err != nil {
		return err
	}

	// start reactor
	reactorSignal := make(chan struct{})
	defer close(reactorSignal)
	go func() {
		defer runtime.HandleCrash()
		instance.reactor.Run(reactorSignal, instance.thread.HandleRequest)
	}()

	runtime.WaitClose(stopCh, instance.shutdown)
	return nil
}

// =====================private methods =================
func (instance *Instance) startListenSchemas(signal <-chan struct{}) error {
	// prepare servers
	for _, schema := range instance.schemas {
		// listen with context and connection register callback function
		if err := schema.Listen(signal, instance.reactor.initializeConnection); err != nil {
			return err
		}

		log.Printf("start listener: %v\n", schema)
	}
	return nil
}

func (instance *Instance) shutdown() {
	log.Println("server shutdown complete")
}
