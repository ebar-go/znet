package codec

import (
	"errors"
	"github.com/ebar-go/ego/utils/binary"
)

// Codec represents a interface that pack/unpack source message
type Codec interface {
	// Pack encode source message by *Packet object
	Pack(packet *Packet, data any) ([]byte, error)

	// Unpack decode source message into *Packet object
	Unpack(packet *Packet, msg []byte) error
	UnpackHeader(buf []byte) (length int)
}

// Options represents codec options
type Options struct {
	// ContentType is data content type
	ContentType int

	headerSize, headerOffset             int
	packetLengthSize, packetLengthOffset int
	operateSize, operateOffset           int
	contentTypeSize, contentTypeOffset   int
	seqSize, seqOffset                   int
}

func defaultOptions() *Options {
	return &Options{
		ContentType:      ContentTypeJSON,
		headerSize:       10,
		packetLengthSize: 4,
		operateSize:      2,
		contentTypeSize:  2,
		seqSize:          2,
	}
}

type DefaultCodec struct {
	options *Options
	endian  binary.Endian
}

type Option func(options *Options)

// Default returns the default codec implementation,the packet is composed by :
// |-------------- header ------------- |-------- body --------|
// |packetLength|operate|contentType|seq|-------- body --------|
// |     4      |   2   |      2    | 2 |          n           |
func Default(opts ...Option) Codec {
	options := defaultOptions()
	for _, setter := range opts {
		setter(options)
	}

	options.headerOffset = options.headerSize
	options.packetLengthOffset = 0 + options.packetLengthSize
	options.operateOffset = options.packetLengthOffset + options.operateSize
	options.contentTypeOffset = options.operateOffset + options.contentTypeSize
	options.seqOffset = options.contentTypeOffset + options.seqSize

	return DefaultCodec{options: options, endian: binary.BigEndian()}
}

func (codec DefaultCodec) Pack(packet *Packet, data any) ([]byte, error) {
	body, err := packet.Marshal(data)
	if err != nil {
		return nil, err
	}

	// packet header and body
	length := len(body) + codec.options.headerSize
	buf := make([]byte, length)

	codec.endian.PutInt32(buf[0:codec.options.packetLengthOffset], int32(length))
	codec.endian.PutInt16(buf[codec.options.packetLengthOffset:codec.options.operateOffset], packet.Operate)
	codec.endian.PutInt16(buf[codec.options.operateOffset:codec.options.contentTypeOffset], packet.ContentType)
	codec.endian.PutInt16(buf[codec.options.contentTypeOffset:codec.options.seqOffset], packet.Seq)
	codec.endian.PutString(buf[codec.options.headerSize:], string(body))
	return buf, nil
}

func (codec DefaultCodec) Unpack(packet *Packet, msg []byte) error {
	if len(msg) < codec.options.headerSize {
		return errors.New("unexpected message")
	}

	length := int(codec.endian.Int32(msg[0:codec.options.packetLengthOffset]))
	packet.Operate = codec.endian.Int16(msg[codec.options.packetLengthOffset:codec.options.operateOffset])
	packet.ContentType = codec.endian.Int16(msg[codec.options.operateOffset:codec.options.contentTypeOffset])
	packet.Seq = codec.endian.Int16(msg[codec.options.contentTypeOffset:codec.options.seqOffset])

	if length > len(msg) {
		return errors.New("unexpected packet length")
	}

	if length > codec.options.headerOffset {
		packet.Body = msg[codec.options.headerOffset:length]
	}

	return nil
}

func (codec DefaultCodec) UnpackHeader(buf []byte) (length int) {
	offset := codec.options.packetLengthOffset
	length = int(codec.endian.Int32(buf[:offset]))
	return

}
