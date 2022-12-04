package main

import (
	"github.com/ebar-go/ego/errors"
	"github.com/ebar-go/ego/utils/convert"
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

	ActionNewUserMessageNotify    = 101
	ActionNewChannelMessageNotify = 102
)

type Handler struct {
	codec    codec.Codec
	users    *structure.ConcurrentMap[string, *znet.Connection]
	channels *structure.ConcurrentMap[string, *Channel]
}

func New() *Handler {
	return &Handler{
		codec:    codec.NewJsonCodec(),
		users:    structure.NewConcurrentMap[string, *znet.Connection](),
		channels: structure.NewConcurrentMap[string, *Channel](),
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
		ID:      "msg" + uuid.NewV4().String(),
		Content: req.Content,
		Sender: User{
			ID:   GetStringFromConnection(ctx.Conn(), "uid"),
			Name: GetStringFromConnection(ctx.Conn(), "name"),
		},
		CreatedAt: time.Now().UnixMilli(),
	}
	p, err := packet.EncodeWith(ActionNewUserMessageNotify, 1, message)

	if err != nil {
		return nil, errors.WithMessage(err, "encode packet")
	}
	if _, err = receiver.Write(p); err != nil {
		return nil, errors.WithMessage(err, "write message")
	}

	resp = &SendUserMessageResponse{ID: message.ID}
	return
}

type Channel struct {
	Name    string `json:"name"`
	Members []string
}

func (handler *Handler) subscribeChannel(ctx *znet.Context, req *SubscribeChannelRequest) (resp *SubscribeChannelResponse, err error) {
	channel, exist := handler.channels.Get(req.Name)
	if !exist {
		channel = &Channel{Name: req.Name, Members: make([]string, 0, 100)}
		channel.Members = append(channel.Members, ctx.Conn().ID())
		handler.channels.Set(req.Name, channel)
		return
	}

	uid := GetStringFromConnection(ctx.Conn(), "uid")
	for _, member := range channel.Members {
		if member == uid {
			return
		}
	}

	channel.Members = append(channel.Members, uid)

	return
}

func (handler *Handler) sendChannelMessage(ctx *znet.Context, req *SendChannelMessageRequest) (resp *SendChannelMessageResponse, err error) {
	channel, err := handler.channels.Find(req.Channel)
	if err != nil {
		return nil, errors.WithMessage(err, "get channel")
	}

	packet := codec.NewPacket(handler.codec)

	message := ChannelMessage{
		Message: Message{
			ID:      "msg" + uuid.NewV4().String(),
			Content: req.Content,
			Sender: User{
				ID:   GetStringFromConnection(ctx.Conn(), "uid"),
				Name: GetStringFromConnection(ctx.Conn(), "name"),
			},
			CreatedAt: time.Now().UnixMilli(),
		},
		Channel: channel.Name,
	}
	p, err := packet.EncodeWith(ActionNewChannelMessageNotify, 1, message)

	if err != nil {
		return nil, errors.WithMessage(err, "encode packet")
	}

	for _, member := range channel.Members {
		receiver, err := handler.users.Find(member)
		if err != nil {
			continue
		}
		if _, err = receiver.Write(p); err != nil {
			continue
		}
	}

	resp = &SendChannelMessageResponse{ID: message.ID}
	return
}

func (handler *Handler) queryHistoryMessage(ctx *znet.Context, req *QueryHistoryMessageRequest) (resp *QueryHistoryMessageResponse, err error) {
	return
}

func GetStringFromConnection(conn *znet.Connection, key string) string {
	val, ok := conn.Property().Get(key)
	if !ok {
		return ""
	}
	return convert.ToString(val)
}
