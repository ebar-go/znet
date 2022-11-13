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

	DecodeBytes(bytes []byte) (buf []byte, err error)
}
type LineBasedFrameDecoder struct {
	packetLengthSize int
	endian           binary.Endian
}

func NewDecoder(packetLengthSize int) Decoder {
	return &LineBasedFrameDecoder{
		packetLengthSize: packetLengthSize,
		endian:           defaultEndian,
	}
}

func (decoder *LineBasedFrameDecoder) Decode(reader io.Reader) (buf []byte, err error) {
	p := pool.GetByte(decoder.packetLengthSize)
	defer pool.PutByte(p)
	_, err = io.ReadFull(reader, p)
	if err != nil {
		return
	}
	length := int(decoder.endian.Int32(p)) - decoder.packetLengthSize
	if length <= 0 {
		// when connection is closed, first read packet length may be successfully, but connection has closed
		err = errors.New("packet exceeded, connection may be closed")
		return
	}
	buf = make([]byte, length)
	_, err = io.ReadFull(reader, buf)
	return
}

func (decoder *LineBasedFrameDecoder) DecodeBytes(bytes []byte) (buf []byte, err error) {
	length := int(decoder.endian.Int32(bytes[:decoder.packetLengthSize]))
	if length <= 0 || length > len(bytes)-decoder.packetLengthSize {
		err = errors.New("packet exceeded, connection may be closed")
		return
	}

	return bytes[decoder.packetLengthSize:length], nil
}

func (decoder *LineBasedFrameDecoder) Encode(writer io.Writer, buf []byte) (n int, err error) {
	length := decoder.packetLengthSize + len(buf)
	p := pool.GetByte(length)
	defer pool.PutByte(p)
	decoder.endian.PutInt32(p[:decoder.packetLengthSize], int32(length))
	copy(p[decoder.packetLengthSize:], buf)
	return writer.Write(p)
}
