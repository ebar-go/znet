package codec

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestPacket_Unmarshal(t *testing.T) {
	type fields struct {
		ContentType int16
		Body        []byte
	}
	type args struct {
		data any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		err    error
	}{
		{
			name: "json.Unmarshal",
			fields: fields{
				ContentType: ContentTypeJSON,
				Body:        []byte(`{"foo":"bar"}`),
			},
			args: args{
				data: &map[string]any{},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := Packet{
				Operate:     0,
				ContentType: tt.fields.ContentType,
				Seq:         0,
				Body:        tt.fields.Body,
			}
			data := tt.args.data
			assert.Equal(t, tt.err, packet.Unmarshal(data))
		})
	}
}

func TestPacket_Marshal(t *testing.T) {
	type fields struct {
		ContentType int16
	}
	type args struct {
		data any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
		err    error
	}{
		{
			name: "json.Marshal",
			fields: fields{
				ContentType: ContentTypeJSON,
			},
			args: args{
				data: map[string]any{"foo": "bar"},
			},
			want: nil,
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := Packet{
				ContentType: tt.fields.ContentType,
			}
			_, err := packet.Marshal(tt.args.data)
			assert.Equal(t, tt.err, err)
			//assert.Equalf(t, tt.want, got, "Marshal(%v)", tt.args.data)
		})
	}
}

func TestBytes(t *testing.T) {
	buf := []byte("hello world")
	log.Printf("%p", buf)
	go func(a []byte) {
		log.Printf("%p", a)
	}(buf)
	time.Sleep(time.Second)
}
