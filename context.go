package znet

import (
	"context"
	"github.com/ebar-go/znet/codec"
	"github.com/ebar-go/znet/internal"
	"log"
	"math"
)

const (
	maxIndex = math.MaxInt8 / 2
)

// HandleFunc represents a handler function for Context
type HandleFunc func(ctx *Context)

// Context represents a context for request
type Context struct {
	context.Context
	index int8

	engine *ContextEngine
	conn   *Connection

	packet   *codec.Packet
	response any
}

func (ctx *Context) Packet() *codec.Packet {
	return ctx.packet
}

// Conn return instance of Connection
func (ctx *Context) Conn() *Connection {
	return ctx.conn
}

// Next invoke next handler
func (ctx *Context) Next() {
	if ctx.index < maxIndex {
		ctx.index++
		ctx.engine.invoke(ctx, ctx.index)
	}
}

// Abort stop invoke handler
func (ctx *Context) Abort() {
	ctx.index = maxIndex
	log.Printf("[%s] context aborted\n", ctx.Conn().ID())
}

// reset clear the context properties
func (ctx *Context) reset(conn *Connection, packet *codec.Packet) {
	ctx.index = 0
	ctx.conn = conn
	ctx.Context = context.Background()
	ctx.response = nil
	ctx.packet = packet
}

type ContextEngine struct {
	handleChains []HandleFunc // is a list of handlers

	provider internal.Provider[*Context] // is a pool for Context
}

func NewContextEngine() *ContextEngine {
	eng := &ContextEngine{}
	eng.provider = internal.NewSyncPoolProvider[*Context](func() interface{} {
		return &Context{engine: eng}
	})
	return eng
}

func (e *ContextEngine) AcquireAndResetContext(conn *Connection, packet *codec.Packet) *Context {
	ctx := e.AcquireContext()
	ctx.reset(conn, packet)
	return ctx
}

func (e *ContextEngine) AcquireContext() *Context {
	return e.provider.Acquire()
}

func (e *ContextEngine) ReleaseContext(ctx *Context) {
	e.provider.Release(ctx)
}

// invokeContextHandler invoke context handler chain
func (e *ContextEngine) invoke(ctx *Context, index int8) {
	if int(index) > len(e.handleChains)-1 {
		return
	}
	e.handleChains[index](ctx)
}
