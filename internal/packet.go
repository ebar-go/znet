package internal

import (
	"encoding/json"
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

	return proto.Marshal(data.(proto.Message))
}

func (packet Packet) Unmarshal(data any) error {
	if packet.ContentType == ContentTypeJSON {
		return json.Unmarshal(packet.Body, data)
	}

	return proto.Unmarshal(packet.Body, data.(proto.Message))
}
