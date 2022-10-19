package znet

import (
	"github.com/ebar-go/ego/utils/runtime"
)

// Engine represents context manager
type Engine struct {
	handleChains    []HandleFunc
	contextProvider ContextProvider
}

// Use registers middleware
func (e *Engine) Use(handler ...HandleFunc) {
	e.handleChains = append(e.handleChains, handler...)
}

// AcquireContext acquire context
func (e *Engine) AcquireContext() *Context {
	return e.contextProvider.AcquireContext()
}

// HandleContext handles context
func (e *Engine) HandleContext(ctx *Context) {
	defer func() {
		runtime.HandleCrash()
		// release Context
		e.contextProvider.ReleaseContext(ctx)
	}()

	e.invokeContextHandler(ctx, 0)

}

// ------------------------private methods------------------------

// invokeContextHandler invoke context handler chain
func (e *Engine) invokeContextHandler(ctx *Context, index int8) {
	e.handleChains[index](ctx)
}

func NewEngine() *Engine {
	engine := &Engine{}
	engine.contextProvider = NewSyncPoolContextProvider(func() interface{} {
		return &Context{engine: engine}
	})
	return engine
}
