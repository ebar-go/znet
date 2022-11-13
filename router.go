package znet

import (
	"github.com/ebar-go/znet/internal"
	"github.com/pkg/errors"
)

// Handler is a handler for operation
type Handler func(ctx *Context) (any, error)

// Action isa generic function that is friendly to user
type Action[Request, Response any] func(ctx *Context, request *Request) (*Response, error)

// StandardHandler is a function to convert standard handler.
func StandardHandler[Request, Response any](action Action[Request, Response]) Handler {
	return func(ctx *Context) (any, error) {
		request := new(Request)
		if err := ctx.Packet().Decode(request); err != nil {
			return nil, err
		}
		return action(ctx, request)
	}
}

// Router represents router instance
type Router struct {
	handlers        *internal.Container[int16, Handler]
	errorHandler    func(ctx *Context, err error)
	notFoundHandler HandleFunc
}

func NewRouter() *Router {
	return &Router{
		handlers:        internal.NewContainer[int16, Handler](),
		errorHandler:    nil,
		notFoundHandler: nil,
	}
}

// Route register handler for operate
func (router *Router) Route(operate int16, handler Handler) *Router {
	router.handlers.Set(operate, handler)
	return router
}

// OnNotFound is called when operation is not found
func (router *Router) OnNotFound(handler HandleFunc) *Router {
	router.notFoundHandler = handler
	return router
}

// OnError is called when an error is encountered while processing a request
func (router *Router) OnError(handler func(ctx *Context, err error)) *Router {
	router.errorHandler = handler
	return router
}

//==================private methods================

func (router *Router) handleRequest(ctx *Context) {
	// match handler
	handler, ok := router.handlers.Get(ctx.Packet().Header().Operate)
	if !ok {
		router.handleNotFound(ctx)
		ctx.Abort()
		return
	}

	response, err := handler(ctx)
	if err != nil {
		router.handleError(ctx, errors.WithMessage(err, "handle operation"))
		ctx.Abort()
		return
	}

	ctx.response = response
	ctx.Next()
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
