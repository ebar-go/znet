package codec

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefault(t *testing.T) {
	codec := Default()
	assert.NotNil(t, codec)
}

func TestDefaultCodec_Pack(t *testing.T) {
	codec := Default()
	bytes, err := codec.Pack(&Packet{Operate: 1, ContentType: ContentTypeJSON, Seq: 1}, map[string]any{"foo": "bar"})
	assert.Nil(t, err)
	assert.NotEmpty(t, bytes)
}

func TestDefaultCodec_Unpack(t *testing.T) {
	codec := Default()
	target := &Packet{Operate: 1, ContentType: ContentTypeJSON, Seq: 1}
	bytes, _ := codec.Pack(target, map[string]any{"foo": "bar"})

	packet, err := codec.Unpack(bytes)
	assert.Nil(t, err)
	assert.Equal(t, target.Operate, packet.Operate)
	assert.Equal(t, target.ContentType, packet.ContentType)
	assert.Equal(t, target.Seq, packet.Seq)
}
