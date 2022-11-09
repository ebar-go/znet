package codec

import (
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	"io"
)

type Decoder interface {
	Decode(reader io.Reader, bytes []byte) (n int, err error)
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

func (decoder *LineBasedFrameDecoder) Decode(reader io.Reader, bytes []byte) (n int, err error) {
	n, err = reader.Read(bytes[:decoder.packetLengthSize])
	if err != nil {
		return
	}
	packetLength := int(decoder.endian.Int32(bytes[:decoder.packetLengthSize]))
	if packetLength < decoder.packetLengthSize || packetLength > len(bytes) {
		// when connection is closed, first read packet length may be successfully, but connection has closed
		err = errors.New("packet exceeded, connection may be closed")
		return
	}
	_, err = reader.Read(bytes[decoder.packetLengthSize:packetLength])
	n = packetLength
	return
}
