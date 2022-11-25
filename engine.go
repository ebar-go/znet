package znet

import (
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/znet/codec"
	"sync"
)

// Engine provide context/handler management
type Engine struct {
	handleChains []HandleFunc // is a list of handlers

	once            sync.Once
	contextProvider pool.Provider[*Context] // is a pool for Context
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Use(handlers ...HandleFunc) {
	e.handleChains = append(e.handleChains, handlers...)
}

func (e *Engine) getProvider() pool.Provider[*Context] {
	e.once.Do(func() {
		e.contextProvider = pool.NewSyncPoolProvider[*Context](func() interface{} {
			return &Context{engine: e}
		})
	})
	return e.contextProvider
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
	provider := e.getProvider()
	// acquire context from provider
	ctx := provider.Acquire()
	ctx.reset(conn, packet)
	defer provider.Release(ctx)

	e.invoke(ctx, 0)
}
