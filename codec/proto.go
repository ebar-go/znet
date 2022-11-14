package codec

import (
	"errors"
	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidProtoMessage = errors.New("invalid proto message")
)

type ProtoCodec struct {
	*Options
}

func NewProtoCodec() *ProtoCodec {
	options := defaultOptions()
	options.complete()
	return &ProtoCodec{Options: options}
}

func (codec *ProtoCodec) Unmarshal(p []byte, data any) error {
	message, ok := data.(proto.Message)
	if !ok {
		return ErrInvalidProtoMessage
	}
	return proto.Unmarshal(p, message)
}

func (codec *ProtoCodec) Marshal(data any) ([]byte, error) {
	message, ok := data.(proto.Message)
	if !ok {
		return nil, ErrInvalidProtoMessage
	}
	return proto.Marshal(message)
}

func (codec *ProtoCodec) Unpack(msg []byte) (packet *Packet) {
	packet = &Packet{codec: codec}
	codec.Options.Unpack(packet, msg)
	return
}
