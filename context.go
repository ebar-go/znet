package znet

import (
	"context"
	"github.com/ebar-go/znet/codec"
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

	thread *Thread
	conn   *Connection

	codec    codec.Codec
	request  []byte
	response any
}

// Request return the source request message
func (ctx *Context) Request() []byte {
	return ctx.request
}

// Conn return instance of Connection
func (ctx *Context) Conn() *Connection {
	return ctx.conn
}

// Next invoke next handler
func (ctx *Context) Next() {
	if ctx.index < maxIndex {
		ctx.index++
		ctx.thread.invokeContextHandler(ctx, ctx.index)
	}
}

// Abort stop invoke handler
func (ctx *Context) Abort() {
	ctx.index = maxIndex
	log.Printf("[%s] context aborted\n", ctx.Conn().ID())
}

// reset clear the context properties
func (ctx *Context) reset(conn *Connection, msg []byte) {
	ctx.index = 0
	ctx.conn = conn
	ctx.Context = context.Background()
	ctx.response = nil
	ctx.request = msg
}
