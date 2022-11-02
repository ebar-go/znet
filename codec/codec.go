package codec

import (
	"encoding/json"
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	"google.golang.org/protobuf/proto"
)

const (
	ContentTypeJSON     = 1
	ContentTypeProtobuf = 2
)

type Header struct {
	Operate     int16
	ContentType int16
	Seq         int16
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

func (options *Options) complete() {
	options.headerOffset = options.headerSize
	options.packetLengthOffset = 0 + options.packetLengthSize
	options.operateOffset = options.packetLengthOffset + options.operateSize
	options.contentTypeOffset = options.operateOffset + options.contentTypeSize
	options.seqOffset = options.contentTypeOffset + options.seqSize
}

type Option func(options *Options)

// Default returns the default codec implementation,the packet is composed by :
// |-------------- header ------------- |-------- body --------|
// |packetLength|operate|contentType|seq|-------- body --------|
// |     4      |   2   |      2    | 2 |          n           |
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

type Codec interface {
	Decode(msg []byte) error
	Pack(data any) ([]byte, error)
	Unpack(data any) error
	Header() Header
}

type DefaultCodec struct {
	options *Options
	endian  binary.Endian
	header  Header
	body    []byte
}

func Default(opts ...Option) *DefaultCodec {
	options := defaultOptions()
	for _, setter := range opts {
		setter(options)
	}
	options.complete()

	return &DefaultCodec{options: options, endian: binary.BigEndian()}
}

func NewPacket(header Header, opts ...Option) *DefaultCodec {
	options := defaultOptions()
	for _, setter := range opts {
		setter(options)
	}
	options.complete()

	return &DefaultCodec{header: header, options: options, endian: binary.BigEndian()}
}

func (codec *DefaultCodec) Pack(data any) ([]byte, error) {
	body, err := codec.marshal(data)
	if err != nil {
		return nil, err
	}

	// packet header and body
	length := len(body) + codec.options.headerSize
	buf := make([]byte, length)

	codec.endian.PutInt32(buf[0:codec.options.packetLengthOffset], int32(length))
	codec.endian.PutInt16(buf[codec.options.packetLengthOffset:codec.options.operateOffset], codec.header.Operate)
	codec.endian.PutInt16(buf[codec.options.operateOffset:codec.options.contentTypeOffset], codec.header.ContentType)
	codec.endian.PutInt16(buf[codec.options.contentTypeOffset:codec.options.seqOffset], codec.header.Seq)
	codec.endian.PutString(buf[codec.options.headerSize:], string(body))
	return buf, nil
}

func (codec *DefaultCodec) Decode(msg []byte) error {
	if len(msg) < codec.options.headerSize {
		return errors.New("unexpected message")
	}

	length := int(codec.endian.Int32(msg[0:codec.options.packetLengthOffset]))
	codec.header.Operate = codec.endian.Int16(msg[codec.options.packetLengthOffset:codec.options.operateOffset])
	codec.header.ContentType = codec.endian.Int16(msg[codec.options.operateOffset:codec.options.contentTypeOffset])
	codec.header.Seq = codec.endian.Int16(msg[codec.options.contentTypeOffset:codec.options.seqOffset])

	if length != len(msg) {
		return errors.New("unexpected packet length")
	}
	codec.body = msg[codec.options.headerOffset:]

	return nil
}

func (codec *DefaultCodec) Unpack(data any) error {
	if codec.header.ContentType == ContentTypeJSON {
		return json.Unmarshal(codec.body, data)
	} else if codec.header.ContentType == ContentTypeProtobuf {
		message, ok := data.(proto.Message)
		if !ok {
			return errors.New("unsupported proto object")
		}

		return proto.Unmarshal(codec.body, message)
	}

	return errors.New("unsupported content type")
}

func (codec *DefaultCodec) Header() Header {
	return codec.header
}

// marshal the given data into body by content type
func (codec *DefaultCodec) marshal(data any) ([]byte, error) {
	if codec.header.ContentType == ContentTypeJSON {
		return json.Marshal(data)
	} else if codec.header.ContentType == ContentTypeProtobuf {
		message, ok := data.(proto.Message)
		if !ok {
			return nil, errors.New("unsupported proto object")
		}

		return proto.Marshal(message)
	}

	return nil, errors.New("unsupported content type")
}
