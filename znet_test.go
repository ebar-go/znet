package znet

import (
	"context"
	"github.com/ebar-go/znet/internal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	instance := New()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	instance.Listen(internal.TCP, ":8081")
	instance.Listen(internal.WEBSOCKET, ":8082")

	err := instance.Run(ctx.Done())
	assert.Nil(t, err)

}
