package main

import (
	"github.com/ebar-go/znet/client"
	"github.com/ebar-go/znet/codec"
	"log"
	"testing"
)

func newClient(name string) (*client.Client, error) {
	conn, err := client.DialTCP("localhost:8081") // tcp
	if err != nil {
		return nil, err
	}

	packet := codec.NewPacket(codec.NewJsonCodec())
	p, err := packet.EncodeWith(ActionLogin, 1, &LoginRequest{Name: "foo"})
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(p)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			response := make([]byte, 512)
			n, err := conn.Read(response)
			if err != nil {
				conn.Close()
				return
			}
			log.Println("receive response: ", string(response[:n]))
		}

	}()

	return conn, nil

}
func TestClient(t *testing.T) {

	t.Run("clientA", func(t *testing.T) {
		newClient("clientA")
		select {}
	})
	t.Run("clientB", func(t *testing.T) {
		conn, err := newClient("clientB")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		packet := codec.NewPacket(codec.NewJsonCodec())
		p, _ := packet.EncodeWith(ActionSendUserMessage, 1, &SendUserMessageRequest{
			ReceiverID: "user:df886555-8380-4bee-8e92-5c6c8e24d9c7",
			Content:    "Hello",
		})
		conn.Write(p)
		select {}
	})

}
