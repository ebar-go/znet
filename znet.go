package znet

import (
	"errors"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/acceptor"
	"log"
)

// Instance implements of Instance interface
type Instance struct {
	options   *Options // options for the event loop
	router    *Router  // router for handlers
	reactor   *Reactor // reactor model
	thread    *Thread  //
	callback  *Callback
	acceptors []acceptor.Instance
}

// New returns a new instance
func New(setters ...Option) *Instance {
	options := completeOptions(setters...)

	return &Instance{
		options:  options,
		reactor:  options.NewReactorOrDie(),
		router:   options.NewRouter(),
		thread:   options.NewThread(),
		callback: options.NewCallback(),
	}
}

// ListenTCP listens for tcp connections
func (instance *Instance) ListenTCP(addr string) {
	instance.acceptors = append(instance.acceptors, acceptor.NewAcceptor(
		acceptor.NewTCPSchema(addr),
		instance.options.Acceptor))
}

// ListenWebsocket listens for websocket connections
func (instance *Instance) ListenWebsocket(addr string) {
	instance.acceptors = append(instance.acceptors, acceptor.NewAcceptor(
		acceptor.NewWebSocketSchema(addr),
		instance.options.Acceptor))
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
	if len(instance.acceptors) == 0 {
		return errors.New("there are no acceptor available")
	}

	instance.thread.Use(instance.options.Middlewares...)
	instance.thread.Use(instance.router.handleRequest(instance.callback.onError))

	// start listeners
	listenerSignal := make(chan struct{})
	defer close(listenerSignal)
	if err := instance.startAcceptor(listenerSignal); err != nil {
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
func (instance *Instance) startAcceptor(signal <-chan struct{}) error {
	handler := instance.reactor.initializeConnection(
		instance.callback.onOpen,
		instance.callback.onClose,
	)

	// prepare servers
	for _, item := range instance.acceptors {
		// listen with context and connection register callback function
		go runtime.WaitClose(signal, item.Shutdown)

		if err := item.Listen(handler); err != nil {
			return err
		}
		log.Printf("Start listening:%v\n", item.Schema())

	}
	return nil
}

func (instance *Instance) shutdown() {
	log.Println("server shutdown complete")
}
