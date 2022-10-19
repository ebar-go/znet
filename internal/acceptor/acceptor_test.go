package acceptor

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestNewAcceptor(t *testing.T) {
	acceptor := NewAcceptor(func(conn net.Conn) {})
	assert.NotNil(t, acceptor)
}

func TestAcceptor_SignalAndDone(t *testing.T) {
	acceptor := NewAcceptor(func(conn net.Conn) {})

	signal := acceptor.Signal()
	assert.NotNil(t, signal)

	go func() {
		time.Sleep(time.Second * 3)
		acceptor.Done()
	}()

	<-signal

}
