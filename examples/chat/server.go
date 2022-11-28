package main

import (
	"fmt"
	"github.com/ebar-go/ego/utils/runtime/signal"
	"github.com/ebar-go/znet"
	"log"
)

func main() {
	instance := znet.New()

	instance.ListenTCP(":8081")
	instance.ListenWebsocket(":8082")

	New().Install(instance.Router())

	if err := instance.Run(signal.SetupSignalHandler()); err != nil {
		log.Fatal(err)
	}
}

const (
	ActionLogin = 1
)

type Chat struct{}

func New() *Chat {
	return &Chat{}
}

func (chat *Chat) Install(router *znet.Router) {
	router.Route(ActionLogin, WrapAction(chat.login))
}

func WrapAction[Request, Response any](action func(ctx *znet.Context, req *Request) (*Response, error)) znet.Handler {
	return func(ctx *znet.Context) (any, error) {
		req := new(Request)
		if err := ctx.Bind(req); err != nil {
			return nil, err
		}

		return action(ctx, req)
	}
}
func (chat *Chat) login(ctx *znet.Context, req *LoginRequest) (resp *LoginResponse, err error) {
	resp = &LoginResponse{Reply: fmt.Sprintf("Welcome: %s", req.Name)}
	return
}
