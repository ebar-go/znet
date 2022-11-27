package client

import (
	"context"
	"github.com/ebar-go/znet/codec"
	"github.com/gobwas/ws"
	"net"
)

type Client struct {
	net.Conn
}

func DialTCP(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Client{Conn: codec.NewLengthFieldBasedFromDecoder(conn, 4)}, nil
}

func DialWebSocket(ctx context.Context, addr string) (*Client, error) {
	conn, _, _, err := ws.Dial(ctx, addr)
	if err != nil {
		return nil, err
	}

	return &Client{Conn: codec.NewWebsocketClientDecoder(conn)}, nil
}
