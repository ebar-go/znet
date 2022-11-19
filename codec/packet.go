package codec

import "errors"

type Packet struct {
	options *Options
	codec   Codec

	Action int16
	Seq    int16
	Body   []byte
}

func NewPacket(codec Codec) *Packet {
	options := DefaultOptions()
	return &Packet{codec: codec, options: options}
}

func (p *Packet) Marshal(data any) (err error) {
	p.Body, err = p.codec.Marshal(data)
	return
}

func (p *Packet) Unmarshal(data any) (err error) {
	return p.codec.Unmarshal(p.Body, data)
}

func (p *Packet) Pack() ([]byte, error) {
	options := p.options
	// packet header and body
	length := len(p.Body) + options.headerSize
	buf := make([]byte, length)

	endian := options.endian
	endian.PutInt16(buf[0:options.actionOffset], p.Action)
	endian.PutInt16(buf[options.actionOffset:options.seqOffset], p.Seq)

	copy(buf[options.headerSize:], p.Body)
	return buf, nil
}

func (p *Packet) Unpack(msg []byte) error {
	options := p.options
	if len(msg) < options.headerOffset {
		return errors.New("msg is too short")
	}
	p.Action = options.endian.Int16(msg[0:options.actionOffset])
	p.Seq = options.endian.Int16(msg[options.actionOffset:options.seqOffset])
	p.Body = msg[options.headerOffset:]

	return nil
}
