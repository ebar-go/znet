package znet

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestNewSingleSubReactor(t *testing.T) {
	sub := NewSingleSubReactor(1024)
	assert.NotNil(t, sub)
}

func TestSubReactor_GetConnection(t *testing.T) {
	sub := NewSingleSubReactor(1024)
	assert.Nil(t, sub.GetConnection(1))
}

func TestSubReactor_RegisterConnection(t *testing.T) {
	sub := NewSingleSubReactor(1024)

	fd := 1
	assert.Nil(t, sub.GetConnection(fd))
	sub.RegisterConnection(&Connection{fd: fd})

	conn := sub.GetConnection(fd)
	assert.NotNil(t, conn)
	assert.Equal(t, fd, conn.fd)
}

func TestSubReactor_UnregisterConnection(t *testing.T) {
	sub := NewSingleSubReactor(1024)
	fd := 1
	sub.RegisterConnection(&Connection{fd: fd})

	conn := sub.GetConnection(fd)
	assert.NotNil(t, conn)
	assert.Equal(t, fd, conn.fd)

	// unregister connection
	sub.UnregisterConnection(conn)
	assert.Nil(t, sub.GetConnection(fd))
}

func TestSubReactor_OfferAndPolling(t *testing.T) {
	sub := NewSingleSubReactor(1024)
	sub.RegisterConnection(NewConnection(nil, 1))

	stop := make(chan struct{})
	go sub.Polling(stop, func(fd int) {
		log.Println(fd)
	})

	sub.Offer(1, 2, 3)

	go func() {
		time.Sleep(time.Second * 3)
		close(stop)
	}()
	<-stop
}
