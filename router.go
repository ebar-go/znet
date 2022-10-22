package znet

import (
	"github.com/ebar-go/znet/codec"
	"github.com/ebar-go/znet/internal"
	"github.com/pkg/errors"
	"sync"
)

type Handler func(ctx *Context) (any, error)

// StandardHandler is a function to convert standard handler.
func StandardHandler[Request, Response any](action func(ctx *Context, request *Request) (*Response, error)) Handler {
	return func(ctx *Context) (any, error) {
		request := new(Request)
		if err := ctx.Request().Unmarshal(request); err != nil {
			return nil, err
		}
		return action(ctx, request)
	}
}

// Router represents router instance
type Router struct {
	rwm             sync.RWMutex
	handlers        map[int16]Handler
	codec           codec.Codec
	errorHandler    func(ctx *Context, err error)
	notFoundHandler HandleFunc
	requestHandler  HandleFunc
	packetProvider  internal.Provider[*codec.Packet]
}

// WithCodec is allowed to be used with the given codec, default is codec.DefaultCodec
func (router *Router) WithCodec(codec codec.Codec) *Router {
	router.codec = codec
	return router
}

// Route register handler for operate
func (router *Router) Route(operate int16, handler Handler) *Router {
	router.rwm.Lock()
	router.handlers[operate] = handler
	router.rwm.Unlock()
	return router
}

// OnNotFound is called when operation is not found
func (router *Router) OnNotFound(handler func(ctx *Context)) *Router {
	router.notFoundHandler = handler
	return router
}

// OnError is called when an error is encountered while processing a request
func (router *Router) OnError(handler func(ctx *Context, err error)) *Router {
	router.errorHandler = handler
	return router
}

func (router *Router) unpack(ctx *Context) {
	// new packet instance from pool, release it after finished
	packet := router.packetProvider.Acquire()
	defer router.packetProvider.Release(packet)
	packet.Reset()

	// unpack
	err := router.codec.Unpack(packet, ctx.msg)
	if err != nil {
		router.handleError(ctx, err)
		ctx.Abort()
		return
	}
	ctx.request = packet
	ctx.Next()
}
func (router *Router) handleRequest(ctx *Context) {
	if router.requestHandler != nil {
		router.requestHandler(ctx)
	} else {
		router.onRequest(ctx)
	}
}
func (router *Router) onRequest(ctx *Context) {
	packet := ctx.request
	// match handler
	router.rwm.RLock()
	handler, ok := router.handlers[packet.Operate]
	router.rwm.RUnlock()
	if !ok {
		router.handleNotFound(ctx)
		return
	}

	response, err := handler(ctx)
	if err != nil {
		router.handleError(ctx, errors.WithMessage(err, "handle operation"))
		return
	}

	packet.Seq++
	// pack response
	msg, err := router.codec.Pack(packet, response)
	if err != nil {
		router.handleError(ctx, errors.WithMessage(err, "invalid response"))
		return
	}
	ctx.Conn().Push(msg)
}

func (router *Router) handleError(ctx *Context, err error) {
	if router.errorHandler != nil {
		router.errorHandler(ctx, err)
	}
}
func (router *Router) handleNotFound(ctx *Context) {
	if router.notFoundHandler != nil {
		router.notFoundHandler(ctx)
	}
}

func NewRouter() *Router {
	return &Router{
		handlers:        map[int16]Handler{},
		codec:           codec.Default(),
		errorHandler:    nil,
		notFoundHandler: nil,
		packetProvider: internal.NewSyncPoolProvider[*codec.Packet](func() interface{} {
			return &codec.Packet{}
		}),
	}
}
