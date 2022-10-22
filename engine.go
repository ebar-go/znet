package znet

import (
	"github.com/ebar-go/znet/internal"
)

// Engine represents context manager
type Engine struct {
	// handlerChains is a list of handlers
	handleChains []HandleFunc

	// contextProvider is a provider for context
	contextProvider internal.Provider[*Context]
}

// Use registers middleware
func (e *Engine) Use(handler ...HandleFunc) {
	e.handleChains = append(e.handleChains, handler...)
}

// HandleRequest handles context
func (e *Engine) HandleRequest(conn *Connection, msg []byte) {
	ctx := e.newContext(conn, msg)
	defer e.releaseContext(ctx)

	e.invokeContextHandler(ctx, 0)
}

// ------------------------private methods------------------------

// NewContext return a new Context instance
func (e *Engine) newContext(conn *Connection, bytes []byte) *Context {
	// acquire context from provider
	ctx := e.contextProvider.Acquire()

	// reset stateful properties
	ctx.reset(conn, bytes)
	return ctx
}

// ReleaseContext releases context
func (e *Engine) releaseContext(ctx *Context) {
	e.contextProvider.Release(ctx)
}

// invokeContextHandler invoke context handler chain
func (e *Engine) invokeContextHandler(ctx *Context, index int8) {
	if int(index) > len(e.handleChains)-1 {
		return
	}
	e.handleChains[index](ctx)
}

// NewEngine returns a new Engine instance
func NewEngine() *Engine {
	engine := &Engine{}
	engine.contextProvider = internal.NewSyncPoolProvider[*Context](func() interface{} {
		return &Context{engine: engine}
	})
	return engine
}
