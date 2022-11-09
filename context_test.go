package znet

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestContext_Request(t *testing.T) {

}

func TestContext_Conn(t *testing.T) {
	ctx := &Context{conn: NewConnection(nil, 1)}
	assert.Equal(t, 1, ctx.Conn().fd)
}

func TestContext_Next(t *testing.T) {
	engine := NewThread(defaultThreadOptions())
	engine.Use(func(ctx *Context) {
		log.Println("Testing context")
		ctx.Next()
	}, func(ctx *Context) {
		log.Println("Testing context with next")
	})
	ctx := &Context{
		conn:   NewConnection(nil, 1),
		thread: engine,
	}
	engine.invokeContextHandler(ctx, 0)
}

func TestContext_Abort(t *testing.T) {
	engine := NewThread(defaultThreadOptions())
	engine.Use(func(ctx *Context) {
		log.Println("Testing context")
		ctx.Abort()
	}, func(ctx *Context) {
		log.Println("Aborting context")
	})
	ctx := &Context{
		conn:   NewConnection(nil, 1),
		thread: engine,
	}
	engine.invokeContextHandler(ctx, 0)
}

func TestContext_reset(t *testing.T) {

}
