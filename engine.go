package znet

import (
	"github.com/ebar-go/znet/codec"
	"github.com/ebar-go/znet/internal"
	"sync"
)

// Engine provide context/handler management
type Engine struct {
	handleChains []HandleFunc // is a list of handlers

	once            sync.Once
	contextProvider internal.Provider[*Context] // is a pool for Context
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) getProvider() internal.Provider[*Context] {
	e.once.Do(func() {
		e.contextProvider = internal.NewSyncPoolProvider[*Context](func() interface{} {
			return &Context{engine: e}
		})
	})
	return e.contextProvider
}

func (e *Engine) AcquireAndResetContext(conn *Connection, packet *codec.Packet) *Context {
	ctx := e.AcquireContext()
	ctx.reset(conn, packet)
	return ctx
}

func (e *Engine) AcquireContext() *Context {
	return e.getProvider().Acquire()
}

func (e *Engine) ReleaseContext(ctx *Context) {
	e.getProvider().Release(ctx)
}

// invokeContextHandler invoke context handler chain
func (e *Engine) invoke(ctx *Context, index int8) {
	if int(index) > len(e.handleChains)-1 {
		return
	}
	e.handleChains[index](ctx)
}
