package acceptor

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestNewSchema(t *testing.T) {
	schema := NewSchema(TCP, ":8081")
	assert.Equal(t, TCP, schema.Protocol)
	assert.Equal(t, ":8081", schema.Addr)

	schemaWs := NewSchema(WEBSOCKET, ":8082")
	assert.Equal(t, WEBSOCKET, schemaWs.Protocol)
	assert.Equal(t, ":8082", schemaWs.Addr)

}

func TestSchema_String(t *testing.T) {
	schema := NewSchema(TCP, ":8081")
	assert.Equal(t, "tcp://:8081", schema.String())
}

func TestSchema_Listen(t *testing.T) {
	type fields struct {
		Protocol string
		Addr     string
	}
	type args struct {
		stopCh  <-chan struct{}
		handler func(conn net.Conn)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "tcp",
			fields: fields{
				Protocol: TCP,
				Addr:     ":8081",
			},
			args: args{
				stopCh:  make(chan struct{}),
				handler: func(conn net.Conn) {},
			},
			wantErr: nil,
		},
		{
			name: "ws",
			fields: fields{
				Protocol: WEBSOCKET,
				Addr:     ":8082",
			},
			args: args{
				stopCh:  make(chan struct{}),
				handler: func(conn net.Conn) {},
			},
			wantErr: nil,
		},
		{
			name: "unsupported",
			fields: fields{
				Protocol: "http",
				Addr:     ":8083",
			},
			args: args{
				stopCh:  make(chan struct{}),
				handler: func(conn net.Conn) {},
			},
			wantErr: errors.New("unsupported protocol: " + "http"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := Schema{
				Protocol: tt.fields.Protocol,
				Addr:     tt.fields.Addr,
			}
			assert.Equal(t, tt.wantErr, schema.Listen(tt.args.stopCh, tt.args.handler))
		})
	}
}
