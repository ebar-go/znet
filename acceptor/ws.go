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
	*Acceptor
	options Options
	upgrade ws.Upgrader
}

// Run runs websocket acceptor
func (acceptor *WebsocketAcceptor) Listen(onAccept func(conn net.Conn)) (err error) {
	ln, err := net.Listen("tcp", acceptor.schema.Addr)
	if err != nil {
		return err
	}

	// use multiple cpus to improve performance
	for i := 0; i < 1; i++ {
		go func() {
			defer runtime.HandleCrash()
			acceptor.accept(ln, onAccept)
		}()
	}
	return nil
}

// accept connection
func (acceptor *WebsocketAcceptor) accept(ln net.Listener, onAccept func(conn net.Conn)) {
	for {
		select {
		case <-acceptor.done:
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
			onAccept(codec.NewWebsocketDecoder(conn))
		}

	}
}
