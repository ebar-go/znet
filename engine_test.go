package znet

import (
	"github.com/ebar-go/znet/codec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEngine(t *testing.T) {
	engine := NewEngine()
	assert.NotNil(t, engine)
}

func TestEngine_HandleRequest(t *testing.T) {
	engine := NewEngine()
	msg := []byte("foo")
	engine.Use(func(ctx *Context) {
		assert.Equal(t, msg, ctx.msg)
	})
	engine.HandleRequest(nil, msg)
}

func TestEngine_Use(t *testing.T) {
	engine := NewEngine()
	msg := []byte("foo")
	packet := &codec.Packet{Operate: 1}
	engine.Use(func(ctx *Context) {
		ctx.request = packet
		ctx.Next()
	}, func(ctx *Context) {
		assert.Equal(t, msg, ctx.msg)
		assert.Equal(t, packet.Operate, ctx.Request().Operate)
	})
	engine.HandleRequest(nil, msg)
}
