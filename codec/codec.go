package codec

import (
	"encoding/json"
	"errors"
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

type Codec interface {
	Decode(msg []byte) error
	Pack(data any) ([]byte, error)
	Unpack(data any) error
	Header() Header
}

type DefaultCodec struct {
	options *Options

	header Header
	body   []byte
}

func (codec *DefaultCodec) Pack(data any) ([]byte, error) {
	body, err := codec.marshal(data)
	if err != nil {
		return nil, err
	}

	// packet header and body
	length := len(body) + codec.options.headerSize
	buf := make([]byte, length)

	endian := codec.options.endian
	endian.PutInt32(buf[0:codec.options.packetLengthOffset], int32(length))
	endian.PutInt16(buf[codec.options.packetLengthOffset:codec.options.operateOffset], codec.header.Operate)
	endian.PutInt16(buf[codec.options.operateOffset:codec.options.contentTypeOffset], codec.header.ContentType)
	endian.PutInt16(buf[codec.options.contentTypeOffset:codec.options.seqOffset], codec.header.Seq)
	endian.PutString(buf[codec.options.headerSize:], string(body))
	return buf, nil
}

func (codec *DefaultCodec) Decode(msg []byte) error {
	if len(msg) < codec.options.headerSize {
		return errors.New("unexpected message")
	}
	endian := codec.options.endian
	length := int(endian.Int32(msg[0:codec.options.packetLengthOffset]))
	codec.header.Operate = endian.Int16(msg[codec.options.packetLengthOffset:codec.options.operateOffset])
	codec.header.ContentType = endian.Int16(msg[codec.options.operateOffset:codec.options.contentTypeOffset])
	codec.header.Seq = endian.Int16(msg[codec.options.contentTypeOffset:codec.options.seqOffset])

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
