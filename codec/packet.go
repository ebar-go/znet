package codec

type Packet struct {
	codec  Codec
	Action int16
	Seq    int16
	Body   []byte
}

func NewPacket(codec Codec) *Packet {
	return &Packet{codec: codec}
}

func (p *Packet) Marshal(data any) (err error) {
	p.Body, err = p.codec.Marshal(data)
	return
}

func (p *Packet) Unmarshal(data any) (err error) {
	return p.codec.Unmarshal(p.Body, data)
}

type Codec interface {
	Marshal(data any) ([]byte, error)
	Unmarshal(p []byte, data any) error
	Unpack(msg []byte) (packet *Packet)
	Pack(packet *Packet) ([]byte, error)
}
