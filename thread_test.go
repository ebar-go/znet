package znet

import (
	"github.com/ebar-go/znet/codec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestThread(t *testing.T) {
	instance := NewThread(defaultThreadOptions())
	assert.NotNil(t, instance)
}

func TestThread_UseAndHandleRequest(t *testing.T) {
	instance := NewThread(defaultThreadOptions())
	msg := []byte("foo")
	packet := &codec.Packet{Operate: 1}
	instance.Use(func(ctx *Context) {
		ctx.request = packet
		ctx.Next()
	}, func(ctx *Context) {
		assert.Equal(t, msg, ctx.msg)
		assert.Equal(t, packet.Operate, ctx.Request().Operate)
	})
	instance.HandleRequest(nil)
}
