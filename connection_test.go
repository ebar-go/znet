package znet

import (
	"errors"
	"github.com/ebar-go/ego/utils/binary"
	"github.com/ebar-go/znet/internal"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"testing"
	"time"
)

func provideNetConn() net.Conn {
	stop := make(chan struct{})
	go internal.NewSchema(internal.TCP, ":8081").
		Listen(stop, func(conn net.Conn) {
			content := time.Now().String()
			buf := make([]byte, len(content)+4)
			binary.BigEndian().PutInt32(buf[:4], int32(len(content))+4)
			binary.BigEndian().PutString(buf[4:], content)
			conn.Write(buf)
		})

	time.Sleep(time.Second * 1)
	conn, err := net.Dial("tcp", ":8081")
	if err != nil {
		panic(err)
	}
	return conn
}
func TestNewConnection(t *testing.T) {
	connection := NewConnection(provideNetConn(), 1)
	assert.NotNil(t, connection)
	assert.NotEmpty(t, connection.ID())
}

func TestConnection_Close(t *testing.T) {
	connection := NewConnection(provideNetConn(), 1)
	connection.AddBeforeCloseHook(func(conn *Connection) {
		log.Println("called before closed")
	})
	connection.Close()

	// close again
	connection.Close()
}

func TestConnection_Property(t *testing.T) {
	connection := NewConnection(provideNetConn(), 1)
	connection.Property().Set("foo", "bar")

	item, exist := connection.Property().Get("foo")
	assert.True(t, exist)
	assert.Equal(t, "bar", item)
}

func TestConnection_Push(t *testing.T) {
	connection := NewConnection(provideNetConn(), 1)
	connection.Push([]byte("foo"))

	p := make([]byte, 512)
	n, err := connection.Read(p)
	assert.Nil(t, err)
	log.Println("receive:", string(p[:n]))
}

func TestConnection_ReadPacket(t *testing.T) {
	type fields struct {
		content              string
		size                 int
		closeAfterDiaSuccess bool
	}
	type args struct {
		port         string
		packetLength int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantN   int
		wantErr error
	}{
		{
			name: "zero packet length",
			fields: fields{
				content: "foo",
				size:    3,
			},
			args: args{
				port:         ":9081",
				packetLength: 0,
			},
			wantN:   0,
			wantErr: nil,
		}, {
			name: "four packet length",
			fields: fields{
				content: "foo",
				size:    7,
			},
			args: args{
				port:         ":9082",
				packetLength: 4,
			},
			wantN:   0,
			wantErr: nil,
		}, {
			name: "close after connected",
			fields: fields{
				content: "foo",
				size:    20,
			},
			args: args{
				port:         ":9083",
				packetLength: 4,
			},
			wantN:   0,
			wantErr: errors.New("packet exceeded"),
		}, {
			name: "four packet length",
			fields: fields{
				content:              "foo",
				size:                 7,
				closeAfterDiaSuccess: true,
			},
			args: args{
				port:         ":9084",
				packetLength: 4,
			},
			wantN:   0,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stop := make(chan struct{})
			err := internal.NewSchema(internal.TCP, tt.args.port).
				Listen(stop, func(conn net.Conn) {
					time.Sleep(time.Second * 1)
					buf := make([]byte, 10)
					_, err := NewConnection(conn, 1).ReadPacket(buf, tt.args.packetLength)
					//assert.Equal(t, tt.wantErr, err)
					if err != tt.wantErr {
						log.Println("got error: ", err)
					}
				})

			assert.Nil(t, err)
			go func() {
				time.Sleep(time.Second * 3)
				close(stop)
			}()

			conn, err := net.Dial("tcp", tt.args.port)
			assert.Nil(t, err)
			if tt.fields.closeAfterDiaSuccess {
				conn.Close()
				time.Sleep(time.Second)
				return
			}

			if tt.args.packetLength == 0 {
				conn.Write([]byte(tt.fields.content))
			} else {
				buf := make([]byte, tt.fields.size)
				binary.BigEndian().PutInt32(buf[:tt.args.packetLength], int32(len(buf)))
				binary.BigEndian().PutString(buf[tt.args.packetLength:], tt.fields.content)

				conn.Write(buf)
			}

			<-stop

		})
	}
}
