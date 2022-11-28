package main

import (
	"github.com/ebar-go/ego/errors"
	"github.com/ebar-go/ego/utils/runtime/signal"
	"github.com/ebar-go/ego/utils/structure"
	"github.com/ebar-go/znet"
	"github.com/ebar-go/znet/codec"
	uuid "github.com/satori/go.uuid"
	"log"
	"time"
)

func main() {
	instance := znet.New(func(options *znet.Options) {
		options.OnError = func(ctx *znet.Context, err error) {
			log.Printf("[%s]OnError: %v", ctx.Conn().ID(), err)
		}
	})

	instance.ListenTCP(":8081")
	instance.ListenWebsocket(":8082")

	New().Install(instance.Router())

	if err := instance.Run(signal.SetupSignalHandler()); err != nil {
		log.Fatal(err)
	}
}

const (
	ActionLogin               = 1
	ActionSendUserMessage     = 2
	ActionSubscribeChannel    = 3
	ActionSendChannelMessage  = 4
	ActionQueryHistoryMessage = 5
)

type Handler struct {
	codec codec.Codec
	users *structure.ConcurrentMap[string, *znet.Connection]
}

func New() *Handler {
	return &Handler{
		codec: codec.NewJsonCodec(),
		users: structure.NewConcurrentMap[string, *znet.Connection](),
	}
}

func (chat *Handler) Install(router *znet.Router) {
	router.Route(ActionLogin, znet.StandardHandler(chat.login))
	router.Route(ActionSendUserMessage, znet.StandardHandler(chat.sendUserMessage))
	router.Route(ActionSubscribeChannel, znet.StandardHandler(chat.subscribeChannel))
	router.Route(ActionSendChannelMessage, znet.StandardHandler(chat.sendChannelMessage))
	router.Route(ActionQueryHistoryMessage, znet.StandardHandler(chat.queryHistoryMessage))
}

func (handler *Handler) login(ctx *znet.Context, req *LoginRequest) (resp *LoginResponse, err error) {
	uid := uuid.NewV4().String()
	ctx.Conn().Property().Set("uid", uid)
	ctx.Conn().Property().Set("name", req.Name)
	handler.users.Set(uid, ctx.Conn())

	resp = &LoginResponse{ID: uid}
	return
}

func (handler *Handler) sendUserMessage(ctx *znet.Context, req *SendUserMessageRequest) (resp *SendUserMessageResponse, err error) {
	receiver, err := handler.users.Find(req.ReceiverID)
	if err != nil {
		return nil, errors.WithMessage(err, "find receiver")
	}

	packet := codec.NewPacket(handler.codec)

	message := Message{
		ID:        "msg" + uuid.NewV4().String(),
		Content:   req.Content,
		CreatedAt: time.Now().UnixMilli(),
	}
	p, err := packet.EncodeWith(ActionSendUserMessage, 1, message)

	if err != nil {
		return nil, errors.WithMessage(err, "encode packet")
	}
	if _, err = receiver.Write(p); err != nil {
		return nil, errors.WithMessage(err, "write message")
	}

	resp = &SendUserMessageResponse{ID: message.ID}
	return
}

func (handler *Handler) subscribeChannel(ctx *znet.Context, req *SubscribeChannelRequest) (resp *SubscribeChannelResponse, err error) {
	return
}

func (handler *Handler) sendChannelMessage(ctx *znet.Context, req *SendChannelMessageRequest) (resp *SendChannelMessageResponse, err error) {
	return
}

func (handler *Handler) queryHistoryMessage(ctx *znet.Context, req *QueryHistoryMessageRequest) (resp *QueryHistoryMessageResponse, err error) {
	return
}
