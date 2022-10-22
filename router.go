package znet

import (
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
	errorHandler    func(ctx *Context, err error)
	notFoundHandler HandleFunc
	requestHandler  HandleFunc
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

func (router *Router) handleRequest(ctx *Context) {
	packet := ctx.request
	// match handler
	router.rwm.RLock()
	handler, ok := router.handlers[packet.Operate]
	router.rwm.RUnlock()
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

func NewRouter() *Router {
	return &Router{
		handlers:        map[int16]Handler{},
		errorHandler:    nil,
		notFoundHandler: nil,
	}
}
