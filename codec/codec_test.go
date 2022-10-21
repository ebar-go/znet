package codec

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefault(t *testing.T) {
	codec := Default()
	assert.NotNil(t, codec)

	codecWithOptions := Default(func(options *Options) {
		options.headerSize = 30
	})
	assert.NotNil(t, codecWithOptions)
}

func TestDefaultCodec_Pack(t *testing.T) {
	codec := Default()
	bytes, err := codec.Pack(&Packet{Operate: 1, ContentType: ContentTypeJSON, Seq: 1}, map[string]any{"foo": "bar"})
	assert.Nil(t, err)
	assert.NotEmpty(t, bytes)

	x := map[string]interface{}{
		"foo": make(chan int),
	}
	bytes, err = codec.Pack(&Packet{Operate: 2, ContentType: ContentTypeJSON, Seq: 2}, x)
	assert.NotNil(t, err)
	assert.Empty(t, bytes)
}

func TestDefaultCodec_Unpack(t *testing.T) {
	codec := Default()
	target := &Packet{Operate: 1, ContentType: ContentTypeJSON, Seq: 1}
	bytes, _ := codec.Pack(target, map[string]any{"foo": "bar"})

	packet := &Packet{}
	err := codec.Unpack(packet, bytes)
	assert.Nil(t, err)
	assert.Equal(t, target.Operate, packet.Operate)
	assert.Equal(t, target.ContentType, packet.ContentType)
	assert.Equal(t, target.Seq, packet.Seq)

	err = codec.Unpack(packet, []byte("foo"))
	assert.NotNil(t, err)
}
