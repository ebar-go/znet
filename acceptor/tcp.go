package acceptor

import (
	"context"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/codec"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"syscall"
)

// TCPAcceptor represents tcp acceptor
type TCPAcceptor struct {
	*Acceptor
	options Options
}

// Run runs the acceptor
func (acceptor *TCPAcceptor) Listen(onAccept func(conn net.Conn)) (err error) {
	if acceptor.options.ReusePort {
		return acceptor.listenReuseAddress(onAccept)
	}
	addr, err := net.ResolveTCPAddr("tcp", acceptor.schema.Addr)
	if err != nil {
		return errors.WithMessage(err, "resolve tcp addr")
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}

	acceptor.serve(listener, onAccept)

	return
}

func (acceptor *TCPAcceptor) listenReuseAddress(onAccept func(conn net.Conn)) (err error) {
	var cfg net.ListenConfig
	cfg.Control = func(network, address string, c syscall.RawConn) error {
		return c.Control(func(fd uintptr) {
			syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEADDR, 1)
			syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
		})
	}

	for i := 0; i < acceptor.options.reuseThread; i++ {
		listener, lastErr := cfg.Listen(context.Background(), "tcp", acceptor.schema.Addr)
		if lastErr != nil {
			return lastErr
		}

		acceptor.serve(listener.(*net.TCPListener), onAccept)
	}
	return
}

func (acceptor *TCPAcceptor) serve(lis *net.TCPListener, onAccept func(conn net.Conn)) {
	for i := 0; i < acceptor.options.Core; i++ {
		go func() {
			defer runtime.HandleCrash()
			acceptor.accept(lis, onAccept)
		}()
	}
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
