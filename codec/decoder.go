package codec

import (
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	"github.com/ebar-go/ego/utils/pool"
	"io"
)

type Decoder interface {
	Decode(reader io.Reader) (buf []byte, err error)
	Encode(writer io.Writer, buf []byte) (n int, err error)
}

// LengthFieldBasedFrameDecode implements Decoder interface by decode length field
type LengthFieldBasedFrameDecode struct {
	offset int
	endian binary.Endian
}

func NewDecoder(offset int) Decoder {
	return &LengthFieldBasedFrameDecode{
		offset: offset,
		endian: defaultEndian,
	}
}

func (decoder *LengthFieldBasedFrameDecode) Decode(reader io.Reader) (buf []byte, err error) {
	// read length field of packet
	p := pool.GetByte(decoder.offset)
	defer pool.PutByte(p)
	_, err = io.ReadFull(reader, p)
	if err != nil {
		return
	}

	// read other part of packet
	length := int(decoder.endian.Int32(p)) - decoder.offset
	if length <= 0 {
		// when connection is closed, first read packet length may be successfully, but connection has closed
		err = errors.New("packet exceeded, connection may be closed")
		return
	}
	buf = make([]byte, length)
	_, err = io.ReadFull(reader, buf)
	return
}

func (decoder *LengthFieldBasedFrameDecode) Encode(writer io.Writer, buf []byte) (n int, err error) {
	length := decoder.offset + len(buf)
	p := pool.GetByte(length)
	defer pool.PutByte(p)
	decoder.endian.PutInt32(p[:decoder.offset], int32(length))
	copy(p[decoder.offset:], buf)
	return writer.Write(p)
}
