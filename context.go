package znet

import (
	"context"
	"github.com/ebar-go/znet/internal/codec"
	"log"
	"math"
)

const (
	maxIndex = math.MaxInt8 / 2
)

// HandlerFunc represents a handler function for Context
type HandleFunc func(ctx *Context)

// Context represents a context for request
type Context struct {
	context.Context
	engine  *Engine
	conn    *Connection
	body    []byte
	index   int8
	request *codec.Packet
}

// Request return the request packet
func (ctx *Context) Request() *codec.Packet {
	return ctx.request
}

// Conn return instance of Connection
func (ctx *Context) Conn() *Connection {
	return ctx.conn
}

// RawBody returns request raw body
func (ctx *Context) RawBody() []byte {
	return ctx.body
}

// Next invoke next handler
func (ctx *Context) Next() {
	if ctx.index < maxIndex {
		ctx.index++
		ctx.engine.invokeContextHandler(ctx, ctx.index)
	}
}

// Abort stop invoke handler
func (ctx *Context) Abort() {
	ctx.index = maxIndex
	log.Printf("[%s] context aborted\n", ctx.Conn().ID())
}

// reset clear the context properties
func (ctx *Context) reset(conn *Connection, body []byte) {
	ctx.index = 0
	ctx.body = body
	ctx.conn = conn
	ctx.Context = context.Background()
}
