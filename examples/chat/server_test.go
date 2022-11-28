package main

import (
	"github.com/ebar-go/znet/client"
	"github.com/ebar-go/znet/codec"
	"log"
	"testing"
)

func TestClient(t *testing.T) {
	conn, err := client.DialTCP("localhost:8081") // tcp
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	packet := codec.NewPacket(codec.NewJsonCodec())
	p, err := packet.EncodeWith(ActionLogin, 1, &LoginRequest{Name: "foo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = conn.Write(p)
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	response := make([]byte, 512)
	n, err := conn.Read(response)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	log.Println("receive response: ", string(response[:n]))

}
