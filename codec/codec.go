package codec

type Codec interface {
	Marshal(data any) ([]byte, error)
	Unmarshal(p []byte, data any) error
	Unpack(msg []byte) (packet *Packet)
	Pack(packet *Packet) ([]byte, error)
}
