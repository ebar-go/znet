package codec

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPacket_Marshal(t *testing.T) {
	packet := &Packet{
		Operate:     1,
		ContentType: ContentTypeJSON,
		Seq:         0,
		Body:        nil,
	}
	bytes, err := packet.Marshal(map[string]any{"foo": "bar"})
	assert.Nil(t, err)
	assert.NotEmpty(t, bytes)
}

func TestPacket_Unmarshal(t *testing.T) {
	packet := &Packet{
		Operate:     1,
		ContentType: ContentTypeJSON,
		Seq:         0,
		Body:        []byte(`{"foo": "bar"}`),
	}
	data := make(map[string]any)
	err := packet.Unmarshal(&data)
	assert.Nil(t, err)
	assert.Equal(t, "bar", data["foo"])
}
