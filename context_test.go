package znet

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestContext_Request(t *testing.T) {
	request := []byte("foo")
	ctx := &Context{request: request}
	assert.Equal(t, request, ctx.Request())
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
	engine := NewThread(defaultThreadOptions())
	ctx := &Context{
		conn:    NewConnection(nil, 1),
		thread:  engine,
		Context: context.Background(),
		index:   1,
	}

	ctx.reset(NewConnection(nil, 2), []byte("bar"))
}
