package znet

import (
	"context"
	"github.com/ebar-go/znet/internal"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	instance := New()

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	instance.Listen(internal.TCP, ":8081").
		Listen(internal.WEBSOCKET, ":8082").
		Run(ctx.Done())

}
