package acceptor

import (
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/gobwas/ws"
	"log"
	"net"
)

type WebsocketAcceptor struct {
	base    *Acceptor
	options *Options
	upgrade ws.Upgrader
}

func (acceptor *WebsocketAcceptor) Run(bind string) (err error) {
	ln, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	for i := 0; i < acceptor.options.core; i++ {
		go func() {
			defer runtime.HandleCrash()
			acceptor.accept(ln)
		}()
	}
	return nil
}

func (acceptor *WebsocketAcceptor) Shutdown() {
	acceptor.base.Done()
}

func (acceptor *WebsocketAcceptor) accept(ln net.Listener) {
	for {
		select {
		case <-acceptor.base.Signal():
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("listener.Accept(\"%s\") error(%v)", ln.Addr().String(), err)
				continue
			}

			_, err = acceptor.upgrade.Upgrade(conn)
			if err != nil {
				log.Printf("upgrade(\"%s\") error(%v)", conn.RemoteAddr().String(), err)
				continue
			}
			acceptor.base.handler(conn)
		}

	}
}

func NewWSAcceptor(handler func(conn net.Conn)) *WebsocketAcceptor {
	return &WebsocketAcceptor{
		base: NewAcceptor(handler),
		options: &Options{
			core:            4,
			readBufferSize:  4096,
			writeBufferSize: 4096,
			keepalive:       false,
		},
		upgrade: ws.Upgrader{
			OnHeader: func(key, value []byte) (err error) {
				log.Printf("non-websocket header: %q=%q", key, value)
				return
			},
		},
	}

}
