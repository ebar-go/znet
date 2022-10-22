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
	index int8
	msg   []byte

	thread *Thread
	conn   *Connection

	request  *codec.Packet
	response any

	context.Context
}

// Request return the request packet
func (ctx *Context) Request() *codec.Packet {
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
	ctx.msg = msg
	ctx.conn = conn
	ctx.Context = context.Background()
	ctx.response = nil
}
