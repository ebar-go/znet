package codec

import (
	"encoding/json"
	"errors"
	"google.golang.org/protobuf/proto"
)

const (
	OptionContentTypeJson int16 = 1 << 1
)

type Header struct {
	Operate int16
	Seq     int16
	Options int16
}

func (header Header) IsContentTypeJson() bool {
	return header.Options&OptionContentTypeJson == OptionContentTypeJson
}

type Packet struct {
	codec *Codec

	header Header
	body   []byte
}

func (packet *Packet) Encode(data any) ([]byte, error) {
	body, err := packet.marshal(data)
	if err != nil {
		return nil, err
	}

	packet.body = body
	return packet.codec.Pack(packet)
}

func (packet *Packet) Decode(data any) error {
	if packet.header.IsContentTypeJson() {
		return json.Unmarshal(packet.body, data)
	}
	message, ok := data.(proto.Message)
	if !ok {
		return errors.New("unsupported proto object")
	}

	return proto.Unmarshal(packet.body, message)
}

func (packet *Packet) Header() Header {
	return packet.header
}

func (packet *Packet) Body() []byte {
	return packet.body
}

// marshal the given data into body by content type
func (packet *Packet) marshal(data any) ([]byte, error) {
	if packet.header.IsContentTypeJson() {
		return json.Marshal(data)
	}
	message, ok := data.(proto.Message)
	if !ok {
		return nil, errors.New("unsupported proto object")
	}

	return proto.Marshal(message)
}
