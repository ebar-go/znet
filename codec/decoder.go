package codec

import (
	"context"
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	"github.com/ebar-go/ego/utils/pool"
	"github.com/gobwas/ws/wsutil"
	"github.com/lucas-clemente/quic-go"
	"io"
	"net"
	"syscall"
	"time"
)

type LengthFieldBasedFrameDecoder struct {
	net.Conn
	offset int
	endian binary.Endian
}

func NewLengthFieldBasedFromDecoder(conn net.Conn, offset int) net.Conn {
	return &LengthFieldBasedFrameDecoder{
		Conn:   conn,
		offset: offset,
		endian: defaultEndian,
	}
}

// SyscallConn prepare for epoll
func (c *LengthFieldBasedFrameDecoder) SyscallConn() (syscall.RawConn, error) {
	return c.Conn.(syscall.Conn).SyscallConn()
}
func (decoder *LengthFieldBasedFrameDecoder) Read(bytes []byte) (n int, err error) {
	// read length field of packet
	p := pool.GetByte(decoder.offset)
	defer pool.PutByte(p)
	_, err = io.ReadFull(decoder.Conn, p)
	if err != nil {
		return
	}

	length := int(decoder.endian.Int32(p))
	if length <= decoder.offset || length > len(bytes) {
		err = errors.New("invalid length")
		return
	}
	n, err = io.ReadFull(decoder.Conn, bytes[:length-decoder.offset])
	return
}

func (decoder *LengthFieldBasedFrameDecoder) Write(buf []byte) (n int, err error) {
	length := decoder.offset + len(buf)
	p := pool.GetByte(length)
	defer pool.PutByte(p)
	decoder.endian.PutInt32(p[:decoder.offset], int32(length))
	copy(p[decoder.offset:], buf)
	return decoder.Conn.Write(p)
}

type websocketDecoder struct {
	net.Conn
	isClient bool
}

func NewWebsocketDecoder(conn net.Conn) net.Conn {
	return &websocketDecoder{Conn: conn}
}

func NewWebsocketClientDecoder(conn net.Conn) net.Conn {
	return &websocketDecoder{Conn: conn, isClient: true}
}

// SyscallConn prepare for epoll
func (c *websocketDecoder) SyscallConn() (syscall.RawConn, error) {
	return c.Conn.(syscall.Conn).SyscallConn()
}

func (c *websocketDecoder) Read(p []byte) (n int, err error) {
	var buf []byte
	if c.isClient {
		buf, err = wsutil.ReadServerBinary(c.Conn)
	} else {
		buf, err = wsutil.ReadClientBinary(c.Conn)
	}

	if err != nil {
		return
	}
	n = copy(p, buf)
	return
}

func (c *websocketDecoder) Write(p []byte) (n int, err error) {
	if c.isClient {
		err = wsutil.WriteClientBinary(c.Conn, p)
	} else {
		err = wsutil.WriteServerBinary(c.Conn, p)
	}

	if err != nil {
		return
	}
	n = len(p)
	return
}

type quicDecoder struct {
	conn     quic.Connection
	isClient bool
}

func (decoder *quicDecoder) Read(b []byte) (n int, err error) {
	var stream quic.Stream
	if decoder.isClient {
		stream, err = decoder.conn.OpenStreamSync(context.Background())
	} else {
		stream, err = decoder.conn.AcceptStream(context.Background())
	}
	if err != nil {
		return
	}

	return stream.Read(b)
}

func (decoder *quicDecoder) Write(b []byte) (n int, err error) {
	var stream quic.Stream
	if decoder.isClient {
		stream, err = decoder.conn.OpenStreamSync(context.Background())
	} else {
		stream, err = decoder.conn.AcceptStream(context.Background())
	}

	if err != nil {
		return
	}
	return stream.Write(b)
}

func (decoder *quicDecoder) Close() error {
	return decoder.conn.CloseWithError(0, "")
}

func (decoder *quicDecoder) LocalAddr() net.Addr {
	return decoder.conn.LocalAddr()
}

func (decoder *quicDecoder) RemoteAddr() net.Addr {
	return decoder.conn.RemoteAddr()
}

func (decoder *quicDecoder) SetDeadline(t time.Time) error {
	return nil
}

func (decoder *quicDecoder) SetReadDeadline(t time.Time) error {
	return nil
}

func (decoder *quicDecoder) SetWriteDeadline(t time.Time) error {
	return nil
}

// SyscallConn prepare for epoll
func (decoder *quicDecoder) SyscallConn() (syscall.RawConn, error) {
	return nil, nil
}

func NewQUICDecoder(conn quic.Connection) net.Conn {
	return &quicDecoder{conn: conn}
}

func NewQUICClientDecoder(conn quic.Connection) net.Conn {
	return &quicDecoder{conn: conn, isClient: true}
}
