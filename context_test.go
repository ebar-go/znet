package znet

import (
	"context"
	"github.com/ebar-go/znet/codec"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestContext_Request(t *testing.T) {
	operate := int16(1)
	ctx := &Context{request: &codec.Packet{Operate: operate}}
	assert.Equal(t, operate, ctx.Request().Operate)
}

func TestContext_Conn(t *testing.T) {
	ctx := &Context{conn: NewConnection(nil, 1)}
	assert.Equal(t, 1, ctx.Conn().fd)
}

func TestContext_Next(t *testing.T) {
	engine := NewEngine()
	engine.Use(func(ctx *Context) {
		log.Println("Testing context")
		ctx.Next()
	}, func(ctx *Context) {
		log.Println("Testing context with next")
	})
	ctx := &Context{
		conn:   NewConnection(nil, 1),
		engine: engine,
	}
	engine.invokeContextHandler(ctx, 0)
}

func TestContext_Abort(t *testing.T) {
	engine := NewEngine()
	engine.Use(func(ctx *Context) {
		log.Println("Testing context")
		ctx.Abort()
	}, func(ctx *Context) {
		log.Println("Aborting context")
	})
	ctx := &Context{
		conn:   NewConnection(nil, 1),
		engine: engine,
	}
	engine.invokeContextHandler(ctx, 0)
}

func TestContext_reset(t *testing.T) {
	engine := NewEngine()
	ctx := &Context{
		conn:    NewConnection(nil, 1),
		engine:  engine,
		Context: context.Background(),
		request: &codec.Packet{},
		index:   1,
		msg:     []byte("foo"),
	}

	ctx.reset(NewConnection(nil, 2), []byte("bar"))
}
