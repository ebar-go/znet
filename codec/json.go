package codec

import "encoding/json"

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
