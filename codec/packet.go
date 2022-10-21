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

// Marshal marshals the given data into body by content type
func (packet *Packet) Marshal(data any) ([]byte, error) {
	if packet.ContentType == ContentTypeJSON {
		return json.Marshal(data)
	} else if packet.ContentType == ContentTypeProtobuf {
		message, ok := data.(proto.Message)
		if !ok {
			return nil, errors.New("unsupported proto object")
		}

		return proto.Marshal(message)
	}

	return nil, errors.New("unsupported content type")
}

// Unmarshal parses the body by content type and stores the result
func (packet *Packet) Unmarshal(data any) error {
	if packet.ContentType == ContentTypeJSON {
		return json.Unmarshal(packet.Body, data)
	} else if packet.ContentType == ContentTypeProtobuf {
		message, ok := data.(proto.Message)
		if !ok {
			return errors.New("unsupported proto object")
		}

		return proto.Unmarshal(packet.Body, message)
	}

	return errors.New("unsupported content type")
}

func (packet *Packet) Reset() {
	packet.Operate = 0
	packet.ContentType = ContentTypeJSON
	packet.Seq = 0
	packet.Body = nil
}
