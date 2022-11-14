package codec

import (
	"encoding/json"
	"errors"
	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidProtoMessage = errors.New("invalid proto message")
)

type Codec interface {
	Marshal(data any) ([]byte, error)
	Unmarshal(p []byte, data any) error
}

type JsonCodec struct {
}

func NewJsonCodec() *JsonCodec {
	return &JsonCodec{}
}

func (codec *JsonCodec) Unmarshal(p []byte, data any) error {
	return json.Unmarshal(p, data)
}

func (codec *JsonCodec) Marshal(data any) ([]byte, error) {
	return json.Marshal(data)
}

type ProtoCodec struct {
}

func NewProtoCodec() *ProtoCodec {
	return &ProtoCodec{}
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
