package znet

import (
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/ego/utils/structure"
)

// Handler is a handler for operation
type Handler func(ctx *Context) (any, error)

// Action isa generic function that is friendly to user
type Action[Request, Response any] func(ctx *Context, request *Request) (*Response, error)

// StandardHandler is a function to convert standard handler.
func StandardHandler[Request, Response any](action Action[Request, Response]) Handler {
	return func(ctx *Context) (any, error) {
		request := new(Request)
		if err := ctx.Packet().Unmarshal(request); err != nil {
			return nil, err
		}
		return action(ctx, request)
	}
}

// Router represents router instance
type Router struct {
	handlers        *structure.ConcurrentMap[int16, Handler]
	errorHandler    func(ctx *Context, err error)
	notFoundHandler HandleFunc
}

func NewRouter() *Router {
	return &Router{
		handlers:        structure.NewConcurrentMap[int16, Handler](),
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

// ==================private methods================
func (router *Router) handleRequest(ctx *Context) {
	// match handler
	handler, ok := router.handlers.Get(ctx.Packet().Action)
	if !ok {
		router.triggerNotFoundEvent(ctx)
		ctx.Abort()
		return
	}

	var (
		response any
		msg      []byte
	)

	lastErr := runtime.Call(func() (err error) { // compute
		response, err = handler(ctx)
		return
	}, func() error { // encode
		return ctx.packet.Marshal(response)
	}, func() (err error) { // pack
		msg, err = ctx.packet.Pack()
		return
	})

	if lastErr != nil {
		router.triggerErrorEvent(ctx, lastErr)
		ctx.Abort()
		return
	}

	ctx.Next()

	ctx.Conn().Write(msg)
}

func (router *Router) triggerErrorEvent(ctx *Context, err error) {
	if router.errorHandler != nil {
		router.errorHandler(ctx, err)
	}
}
func (router *Router) triggerNotFoundEvent(ctx *Context) {
	if router.notFoundHandler != nil {
		router.notFoundHandler(ctx)
	}
}
