package acceptor

import (
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	"github.com/ebar-go/ego/utils/pool"
	"github.com/gobwas/ws/wsutil"
	"io"
	"net"
	"syscall"
)

type LengthFieldBasedFrameDecoder struct {
	net.Conn
	offset int
	endian binary.Endian
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
}

// SyscallConn prepare for epoll
func (c *websocketDecoder) SyscallConn() (syscall.RawConn, error) {
	return c.Conn.(syscall.Conn).SyscallConn()
}

func (c *websocketDecoder) Read(p []byte) (n int, err error) {
	buf, err := wsutil.ReadClientBinary(c.Conn)
	if err != nil {
		return
	}
	n = copy(p, buf)
	return
}

func (c *websocketDecoder) Write(p []byte) (n int, err error) {
	err = wsutil.WriteServerBinary(c.Conn, p)
	if err != nil {
		return
	}
	n = len(p)
	return
}
