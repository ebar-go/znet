package acceptor

import (
	"github.com/ebar-go/ego/errors"
	"github.com/ebar-go/ego/utils/binary"
	"github.com/ebar-go/ego/utils/pool"
	"github.com/ebar-go/ego/utils/runtime"
	"io"
	"log"
	"net"
)

// TCPAcceptor represents tcp acceptor
type TCPAcceptor struct {
	base    *Acceptor
	options *Options
}

// Run runs the acceptor
func (server *TCPAcceptor) Run(bind string) (err error) {
	addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return errors.WithMessage(err, "resolve tcp addr")
	}

	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}

	// use multiple cpus to improve performance
	for i := 0; i < server.options.Core; i++ {
		go func() {
			defer runtime.HandleCrash()
			server.accept(lis)
		}()
	}

	return
}

// Shutdown shuts down acceptor
func (acceptor *TCPAcceptor) Shutdown() {
	acceptor.base.Done()
}

// accept connection
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

			acceptor.base.handler(&tcpConnAdapter{Conn: conn, offset: 4, endian: binary.BigEndian()})
		}
	}

}

// NewTCPAcceptor returns a new instance of the TCPAcceptor
func NewTCPAcceptor(options *Options, handler func(conn net.Conn)) *TCPAcceptor {
	return &TCPAcceptor{
		base:    NewAcceptor(handler),
		options: options,
	}
}

type tcpConnAdapter struct {
	net.Conn
	offset int
	endian binary.Endian
}

func (a *tcpConnAdapter) Read(bytes []byte) (n int, err error) {
	// read length field of packet
	p := pool.GetByte(a.offset)
	defer pool.PutByte(p)
	_, err = io.ReadFull(a.Conn, p)

	length := a.endian.Int32(p)
	n, err = io.ReadFull(a.Conn, bytes[:length-4])
	return
}

func (a *tcpConnAdapter) Write(buf []byte) (n int, err error) {
	length := a.offset + len(buf)
	p := pool.GetByte(length)
	defer pool.PutByte(p)
	a.endian.PutInt32(p[:a.offset], int32(length))
	copy(p[a.offset:], buf)
	return a.Conn.Write(p)
}
