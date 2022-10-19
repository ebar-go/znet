package codec

import (
	"encoding/json"
	"errors"
	"github.com/golang/protobuf/proto"
)

type Packet struct {
	Operate     int16
	ContentType int16
	Seq         int16
	Body        []byte
}

const (
	ContentTypeJSON     = 1
	ContentTypeProtobuf = 2
)

func (packet Packet) Marshal(data any) ([]byte, error) {
	if packet.ContentType == ContentTypeJSON {
		return json.Marshal(data)
	}

	message, ok := data.(proto.Message)
	if !ok {
		return nil, errors.New("unsupported proto object")
	}

	return proto.Marshal(message)
}

func (packet Packet) Unmarshal(data any) error {
	if packet.ContentType == ContentTypeJSON {
		return json.Unmarshal(packet.Body, data)
	}

	message, ok := data.(proto.Message)
	if !ok {
		return errors.New("unsupported proto object")
	}

	return proto.Unmarshal(packet.Body, message)
}
