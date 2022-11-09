package codec

import (
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	"sync"
)

// Codec represents codec codec
type Codec struct {
	endian binary.Endian

	headerSize, headerOffset             int
	packetLengthSize, packetLengthOffset int
	operateSize, operateOffset           int
	seqSize, seqOffset                   int
	optionSize, optionOffset             int
}

func (codec *Codec) complete() {
	codec.headerOffset = codec.headerSize
	codec.packetLengthOffset = 0 + codec.packetLengthSize
	codec.operateOffset = codec.packetLengthOffset + codec.operateSize
	codec.seqOffset = codec.operateOffset + codec.seqSize
	codec.optionOffset = codec.seqOffset + codec.optionSize
}

func (codec *Codec) Encode(packet *Packet) ([]byte, error) {
	// packet header and body
	length := len(packet.body) + codec.headerSize
	buf := make([]byte, length)

	endian := codec.endian
	endian.PutInt32(buf[0:codec.packetLengthOffset], int32(length))
	endian.PutInt16(buf[codec.packetLengthOffset:codec.operateOffset], packet.header.Operate)
	endian.PutInt16(buf[codec.operateOffset:codec.seqOffset], packet.header.Seq)
	endian.PutInt16(buf[codec.seqOffset:codec.optionOffset], packet.header.Options)

	copy(buf[codec.headerSize:], packet.body)
	return buf, nil
}

func (codec *Codec) Decode(packet *Packet, msg []byte) (err error) {
	if len(msg) < codec.headerSize {
		return errors.New("unexpected message")
	}
	packet.header.Operate = codec.endian.Int16(msg[codec.packetLengthOffset:codec.operateOffset])
	packet.header.Seq = codec.endian.Int16(msg[codec.operateOffset:codec.seqOffset])
	packet.header.Options = codec.endian.Int16(msg[codec.operateOffset:codec.optionOffset])

	packet.body = msg[codec.headerOffset:]

	return
}

func (codec *Codec) NewPacket(msg []byte) (*Packet, error) {
	packet := &Packet{}
	err := codec.Decode(packet, msg)
	return packet, err
}

func (codec *Codec) NewWithHeader(header Header) *Packet {
	return &Packet{codec: codec, header: header}
}

// Default returns the default codec implementation,the packet is composed by :
// |-------------- header ------------- |-------- body --------|
// |packetLength|operate|contentType|seq|-------- body --------|
// |     4      |   2   |      2    | 2 |          n           |
func defaultCodec() *Codec {
	return &Codec{
		headerSize:       10,
		packetLengthSize: 4,
		operateSize:      2,
		seqSize:          2,
		optionSize:       2,
		endian:           defaultEndian,
	}
}

var codecInstance = struct {
	once     sync.Once
	instance *Codec
}{}

func Factory() *Codec {
	codecInstance.once.Do(func() {
		codecInstance.instance = defaultCodec()
		codecInstance.instance.complete()
	})
	return codecInstance.instance
}

var defaultEndian = binary.BigEndian()

func SetEndian(endian binary.Endian) {
	defaultEndian = endian
}
