package acceptor

import (
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/codec"
	"github.com/gobwas/ws"
	"log"
	"net"
)

// WebsocketAcceptor represents websocket acceptor
type WebsocketAcceptor struct {
	base    *Acceptor
	options *Options
	upgrade ws.Upgrader
}

// Run runs websocket acceptor
func (acceptor *WebsocketAcceptor) Run(bind string) (err error) {
	ln, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	// use multiple cpus to improve performance
	for i := 0; i < 1; i++ {
		go func() {
			defer runtime.HandleCrash()
			acceptor.accept(ln)
		}()
	}
	return nil
}

// Shutdown shuts down acceptor
func (acceptor *WebsocketAcceptor) Shutdown() {
	acceptor.base.Done()
}

// accept connection
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
			acceptor.base.handler(codec.NewWebsocketDecoder(conn))
		}

	}
}

// NewWSAcceptor return a new instance of the WebsocketAcceptor
func NewWSAcceptor(options *Options, handler func(conn net.Conn)) *WebsocketAcceptor {
	return &WebsocketAcceptor{
		base:    NewAcceptor(handler),
		options: options,
		upgrade: ws.Upgrader{
			OnHeader: func(key, value []byte) (err error) {
				//log.Printf("non-websocket header: %q=%q", key, value)
				return
			},
		},
	}

}
