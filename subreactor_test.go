package znet

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewSubReactor(t *testing.T) {
	sub := NewSubReactor(1024)
	assert.NotNil(t, sub)
}

func TestSubReactor_GetConnection(t *testing.T) {
	sub := NewSubReactor(1024)
	assert.Nil(t, sub.GetConnection(1))
}

func TestSubReactor_RegisterConnection(t *testing.T) {
	sub := NewSubReactor(1024)

	fd := 1
	assert.Nil(t, sub.GetConnection(fd))
	sub.RegisterConnection(&Connection{fd: fd})

	conn := sub.GetConnection(fd)
	assert.NotNil(t, conn)
	assert.Equal(t, fd, conn.fd)
}

func TestSubReactor_UnregisterConnection(t *testing.T) {
	sub := NewSubReactor(1024)
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
	sub := NewSubReactor(1024)

	stop := make(chan struct{})
	go sub.Polling(stop, func(active int) {
		fmt.Println("active: ", active)
	})

	sub.Offer(1, 2, 3)

	go func() {
		time.Sleep(time.Second * 3)
		close(stop)
	}()
	<-stop
}