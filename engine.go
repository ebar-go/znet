package znet

import (
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/internal"
)

// Engine represents context manager
type Engine struct {
	handleChains    []HandleFunc
	contextProvider internal.Provider[*Context]
}

// Use registers middleware
func (e *Engine) Use(handler ...HandleFunc) {
	e.handleChains = append(e.handleChains, handler...)
}

// AcquireContext acquire context
func (e *Engine) AcquireContext() *Context {
	return e.contextProvider.Acquire()
}

// HandleContext handles context
func (e *Engine) HandleContext(ctx *Context) {
	defer func() {
		runtime.HandleCrash()
		// release Context
		e.contextProvider.Release(ctx)
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
	engine.contextProvider = internal.NewSyncPoolProvider[*Context](func() interface{} {
		return &Context{engine: engine}
	})
	return engine
}
