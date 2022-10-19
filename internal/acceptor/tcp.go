package acceptor

import (
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/pkg/errors"
	"log"
	"net"
)

type TCPAcceptor struct {
	base    *Acceptor
	options *Options
}

func (server *TCPAcceptor) Run(bind string) (err error) {
	addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return errors.WithMessage(err, "resolve tcp addr")
	}

	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}

	for i := 0; i < server.options.core; i++ {
		go func() {
			defer runtime.HandleCrash()
			server.accept(lis)
		}()
	}

	return
}

func (acceptor *TCPAcceptor) Shutdown() {
	acceptor.base.Done()
}

func (acceptor *TCPAcceptor) accept(lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
	)

	for {
		select {
		case <-acceptor.base.Signal():
			return
		default:
			if conn, err = lis.AcceptTCP(); err != nil {
				// if listener close then return
				log.Printf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
				continue
			}
			if err = conn.SetKeepAlive(acceptor.options.keepalive); err != nil {
				log.Printf("conn.SetKeepAlive() error(%v)", err)
				continue
			}
			if err = conn.SetReadBuffer(acceptor.options.readBufferSize); err != nil {
				log.Printf("conn.SetReadBuffer() error(%v)", err)
				continue
			}
			if err = conn.SetWriteBuffer(acceptor.options.writeBufferSize); err != nil {
				log.Printf("conn.SetWriteBuffer() error(%v)", err)
				continue
			}

			acceptor.base.handler(conn)
		}
	}

}

func NewTCPTCPAcceptor(handler func(conn net.Conn)) *TCPAcceptor {
	return &TCPAcceptor{
		base: NewAcceptor(handler),
		options: &Options{
			core:            4,
			readBufferSize:  4096,
			writeBufferSize: 4096,
			keepalive:       false,
		}}
}
