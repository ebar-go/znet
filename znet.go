package znet

import (
	"errors"
	"github.com/ebar-go/ego/component"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/acceptor"
	"log"
)

// Network socket server master
type Network struct {
	options   *Options // options for the event loop
	router    *Router  // router for handlers
	reactor   *Reactor // reactor model
	thread    *Thread  //
	callback  *Callback
	acceptors []acceptor.Instance
}

// New returns a new instance
func New(setters ...Option) *Network {
	options := completeOptions(setters...)

	return &Network{
		options:  options,
		reactor:  options.NewReactorOrDie(),
		router:   options.NewRouter(),
		thread:   options.NewThread(),
		callback: options.NewCallback(),
	}
}

// ListenTCP listens for tcp connections
func (instance *Network) ListenTCP(addr string) {
	instance.acceptors = append(instance.acceptors, acceptor.NewAcceptor(
		acceptor.NewTCPSchema(addr),
		instance.options.Acceptor))
}

// ListenWebsocket listens for websocket connections
func (instance *Network) ListenWebsocket(addr string) {
	instance.acceptors = append(instance.acceptors, acceptor.NewAcceptor(
		acceptor.NewWebSocketSchema(addr),
		instance.options.Acceptor))
}

// ListenQUIC listens for quic connections
func (instance *Network) ListenQUIC(addr string) {
	instance.acceptors = append(instance.acceptors, acceptor.NewAcceptor(
		acceptor.NewQUICSchema(addr),
		instance.options.Acceptor))
}

// Router return instance of Router
func (instance *Network) Router() *Router {
	return instance.router
}

// Run starts the event-loop
func (instance *Network) Run(stopCh <-chan struct{}) error {
	if err := instance.options.Validate(); err != nil {
		return err
	}
	if len(instance.acceptors) == 0 {
		return errors.New("there are no acceptor available")
	}

	instance.thread.Use(instance.options.Middlewares...)
	instance.thread.Use(instance.router.handleRequest(instance.callback.onError))

	component.Event().Trigger(BeforeServerStart, nil)
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

	component.Event().Trigger(AfterServerStart, nil)

	runtime.WaitClose(stopCh, func() {
		component.Event().Trigger(BeforeServerShutdown, nil)
	}, instance.shutdown)

	return nil
}

// =====================private methods =================
func (instance *Network) startAcceptor(signal <-chan struct{}) error {
	handler := instance.reactor.initializeConnection(
		instance.callback.onOpen,
		instance.callback.onClose,
	)
	unsupportedHandler := instance.reactor.initializeUnSupportedReactorConnection(
		instance.callback.onOpen,
		instance.callback.onClose,
		instance.thread.HandleRequest,
	)

	// prepare servers
	for _, item := range instance.acceptors {
		if item.ReactorSupported() {
			if err := item.Listen(handler); err != nil {
				return err
			}
		} else {
			if err := item.Listen(unsupportedHandler); err != nil {
				return err
			}
		}

		log.Printf("Start listener: %v\n", item.Schema())

		// listen with context and connection register callback function
		go runtime.WaitClose(signal, item.Shutdown)

	}
	return nil
}

func (instance *Network) shutdown() {
	component.Event().Trigger(AfterServerShutdown, nil)
}
