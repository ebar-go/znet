package client

import (
	"context"
	"crypto/tls"
	"github.com/ebar-go/znet/codec"
	"github.com/gobwas/ws"
	"github.com/lucas-clemente/quic-go"
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

func DialQUIC(addr string) (*Client, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	conn, err := quic.DialAddr(addr, tlsConf, nil)
	if err != nil {
		return nil, err
	}

	return &Client{Conn: codec.NewQUICDecoder(conn)}, nil
}
