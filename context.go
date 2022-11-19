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

	engine *Engine
	conn   *Connection

	packet *codec.Packet
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
	log.Printf("[%s] context aborted:%d \n", ctx.Conn().ID(), ctx.index)
	ctx.index = maxIndex
}

// reset clear the context properties
func (ctx *Context) reset(conn *Connection, packet *codec.Packet) {
	ctx.Context = context.Background()
	ctx.index = 0
	ctx.conn = conn
	ctx.packet = packet
}
