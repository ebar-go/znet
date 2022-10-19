package znet

import (
	"context"
	"github.com/ebar-go/znet/internal/codec"
	"log"
	"math"
	"sync"
)

const (
	maxIndex = math.MaxInt8 / 2
)

type HandleFunc func(ctx *Context)

// Context represents a context for request
type Context struct {
	context.Context
	engine *Engine
	conn   *Connection
	body   []byte
	index  int8
	packet *codec.Packet
}

func (ctx *Context) Operate() int16 {
	return ctx.packet.Operate
}

// Bind checks the Content-Type to select a binding engine automatically
func (ctx *Context) Bind(container any) error {
	return ctx.packet.Unmarshal(container)
}

// Conn return instance of Connection
func (ctx *Context) Conn() *Connection {
	return ctx.conn
}

// RawBody returns request raw body
func (ctx *Context) RawBody() []byte {
	return ctx.body
}

func (ctx *Context) Next() {
	if ctx.index < maxIndex {
		ctx.index++
		ctx.engine.invokeContextHandler(ctx, ctx.index)
	}
}
func (ctx *Context) Abort() {
	ctx.index = maxIndex
	log.Printf("[%s] context aborted\n", ctx.Conn().UUID())
}

func (ctx *Context) reset(conn *Connection, body []byte) {
	ctx.index = 0
	ctx.body = body
	ctx.conn = conn
	ctx.Context = context.Background()
}

type ContextProvider interface {
	AcquireContext() *Context
	ReleaseContext(ctx *Context)
}

type SyncPoolContextProvider struct {
	pool *sync.Pool
}

func (provider *SyncPoolContextProvider) AcquireContext() *Context {
	return provider.pool.Get().(*Context)
}

func (provider *SyncPoolContextProvider) ReleaseContext(ctx *Context) {
	provider.pool.Put(ctx)
}

func NewSyncPoolContextProvider(constructor func() interface{}) ContextProvider {
	return &SyncPoolContextProvider{pool: &sync.Pool{New: constructor}}
}
