package znet

import (
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
	notFoundHandler HandleFunc
}

func NewRouter() *Router {
	return &Router{
		handlers:        structure.NewConcurrentMap[int16, Handler](),
		notFoundHandler: nil,
	}
}

// Route register handler for action
func (router *Router) Route(action int16, handler Handler) *Router {
	router.handlers.Set(action, handler)
	return router
}

// OnNotFound is called when operation is not found
func (router *Router) OnNotFound(handler HandleFunc) *Router {
	router.notFoundHandler = handler
	return router
}

// ==================private methods================
func (router *Router) handleRequest(onError func(ctx *Context, err error)) HandleFunc {
	return func(ctx *Context) {
		// match handler
		handler, ok := router.handlers.Get(ctx.Packet().Action)
		if !ok {
			router.triggerNotFoundEvent(ctx)
			return
		}

		response, err := handler(ctx)
		if err != nil {
			onError(ctx, err)
			return
		}

		msg, err := ctx.packet.Encode(response)
		if err != nil {
			onError(ctx, err)
			return
		}

		ctx.Conn().Write(msg)

	}

}

func (router *Router) triggerNotFoundEvent(ctx *Context) {
	if router.notFoundHandler != nil {
		router.notFoundHandler(ctx)
	}
}
