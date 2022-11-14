package codec

import (
	"encoding/json"
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	"google.golang.org/protobuf/proto"
)

// Options represents options options
type Options struct {
	endian binary.Endian

	headerSize, headerOffset int
	actionSize, actionOffset int
	seqSize, seqOffset       int
}

func (options *Options) complete() {
	options.headerOffset = options.headerSize
	options.actionOffset = options.actionSize
	options.seqOffset = options.actionOffset + options.seqSize
}

func (options *Options) Pack(packet *Packet) ([]byte, error) {
	// packet header and body
	length := len(packet.Body) + options.headerSize
	buf := make([]byte, length)

	endian := options.endian
	endian.PutInt16(buf[0:options.actionOffset], packet.Action)
	endian.PutInt16(buf[options.actionOffset:options.seqOffset], packet.Seq)

	copy(buf[options.headerSize:], packet.Body)
	return buf, nil
}

func (options *Options) Unpack(packet *Packet, msg []byte) {

	packet.Action = options.endian.Int16(msg[0:options.actionOffset])
	packet.Seq = options.endian.Int16(msg[options.actionOffset:options.seqOffset])

	packet.Body = msg[options.headerOffset:]

	return
}

// Default returns the default options implementation,the packet is composed by :
// |-------------- header ------------- |-------- body --------|
// |packetLength|action|contentType|seq|-------- body --------|
// |     4      |   2   |      2    | 2 |          n           |
func defaultOptions() *Options {
	return &Options{
		endian:     defaultEndian,
		headerSize: 4,
		actionSize: 2,
		seqSize:    2,
	}
}

type JsonCodec struct {
	*Options
}

func NewJsonCodec() *JsonCodec {
	options := defaultOptions()
	options.complete()
	return &JsonCodec{Options: options}
}

func (codec *JsonCodec) Unmarshal(p []byte, data any) error {
	return json.Unmarshal(p, data)
}

func (codec *JsonCodec) Marshal(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (codec *JsonCodec) Unpack(msg []byte) (packet *Packet) {
	packet = &Packet{codec: codec}
	codec.Options.Unpack(packet, msg)
	return
}

var defaultEndian = binary.BigEndian()

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
		return errors.New("invalid proto message")
	}
	return proto.Unmarshal(p, message)
}

func (codec *ProtoCodec) Marshal(data any) ([]byte, error) {
	message, ok := data.(proto.Message)
	if !ok {
		return nil, errors.New("invalid proto message")
	}
	return proto.Marshal(message)
}

func (codec *ProtoCodec) Unpack(msg []byte) (packet *Packet) {
	packet = &Packet{codec: codec}
	codec.Options.Unpack(packet, msg)
	return
}
