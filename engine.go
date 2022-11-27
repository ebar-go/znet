package znet

import (
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/znet/codec"
)

// Engine provide context/handler management
type Engine struct {
	handleChains []HandleFunc // is a list of handlers

	contextProvider pool.Provider[*Context] // is a pool for Context
}

func NewEngine() *Engine {
	e := &Engine{}
	e.contextProvider = pool.NewSyncPoolProvider[*Context](func() interface{} {
		return &Context{engine: e}
	})
	return e
}

func (e *Engine) Use(handlers ...HandleFunc) {
	e.handleChains = append(e.handleChains, handlers...)
}

// invoke process context handler chain
func (e *Engine) invoke(ctx *Context, index int8) {
	if int(index) > len(e.handleChains)-1 {
		return
	}
	e.handleChains[index](ctx)
}

// compute run invoke function with context
func (e *Engine) compute(conn *Connection, packet *codec.Packet) {
	// acquire context from provider
	ctx := e.contextProvider.Acquire()
	ctx.reset(conn, packet)
	defer e.contextProvider.Release(ctx)

	e.invoke(ctx, 0)
}
