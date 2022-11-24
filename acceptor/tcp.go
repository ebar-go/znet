package acceptor

import (
	"github.com/ebar-go/ego/errors"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/codec"
	"log"
	"net"
)

// TCPAcceptor represents tcp acceptor
type TCPAcceptor struct {
	*Acceptor
	options Options
}

// Run runs the acceptor
func (acceptor *TCPAcceptor) Listen(onAccept func(conn net.Conn)) (err error) {
	addr, err := net.ResolveTCPAddr("tcp", acceptor.schema.Addr)
	if err != nil {
		return errors.WithMessage(err, "resolve tcp addr")
	}

	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}

	// use multiple cpus to improve performance
	for i := 0; i < acceptor.options.Core; i++ {
		go func() {
			defer runtime.HandleCrash()
			acceptor.accept(lis, onAccept)
		}()
	}

	return
}

// accept connection
func (acceptor *TCPAcceptor) accept(lis *net.TCPListener, onAccept func(conn net.Conn)) {
	for {
		select {
		case <-acceptor.done:
			return
		default:
			conn, err := lis.AcceptTCP()
			if err != nil {
				// if listener close then return
				log.Printf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
				continue
			}
			if err = conn.SetKeepAlive(acceptor.options.Keepalive); err != nil {
				log.Printf("conn.SetKeepAlive() error(%v)", err)
				continue
			}
			if err = conn.SetReadBuffer(acceptor.options.ReadBufferSize); err != nil {
				log.Printf("conn.SetReadBuffer() error(%v)", err)
				continue
			}
			if err = conn.SetWriteBuffer(acceptor.options.WriteBufferSize); err != nil {
				log.Printf("conn.SetWriteBuffer() error(%v)", err)
				continue
			}

			onAccept(codec.NewLengthFieldBasedFromDecoder(conn, acceptor.options.LengthOffset))
		}
	}

}
