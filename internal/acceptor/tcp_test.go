package acceptor

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestNewTCPAcceptor(t *testing.T) {
	type args struct {
		options *Options
		handler func(conn net.Conn)
	}
	tests := []struct {
		name string
		args args
		want *TCPAcceptor
	}{
		{
			name: "defaultOptions",
			args: args{
				options: DefaultOptions(),
				handler: func(conn net.Conn) {},
			},
			want: &TCPAcceptor{
				base:    NewAcceptor(func(conn net.Conn) {}),
				options: DefaultOptions(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.args.options, tt.want.options)
			assert.NotNil(t, tt.args.handler, tt.want.base.handler)
		})
	}
}

func TestTCPAcceptor_Run(t *testing.T) {
	acceptor := NewTCPAcceptor(DefaultOptions(), func(conn net.Conn) {})

	err := acceptor.Run(":8080")
	assert.Nil(t, err)

	errAcceptor := NewTCPAcceptor(DefaultOptions(), func(conn net.Conn) {})
	err = errAcceptor.Run("invalid:8080")
	assert.NotNil(t, err)

	repeatPortAcceptor := NewTCPAcceptor(DefaultOptions(), func(conn net.Conn) {})
	err = repeatPortAcceptor.Run(":8080")
	assert.NotNil(t, err)

	time.Sleep(time.Second * 3)
	acceptor.Shutdown()
}

func TestTCPAcceptor_Shutdown(t *testing.T) {
	acceptor := NewTCPAcceptor(DefaultOptions(), func(conn net.Conn) {})

	err := acceptor.Run(":9091")
	assert.Nil(t, err)
	time.Sleep(time.Second * 3)
	acceptor.Shutdown()
}
