package znet

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewThread(t *testing.T) {
	thread := NewThread(1024)
	assert.NotNil(t, thread)
}

func TestThread_GetConnection(t *testing.T) {
	thread := NewThread(1024)
	assert.Nil(t, thread.GetConnection(1))
}

func TestThread_RegisterConnection(t *testing.T) {
	thread := NewThread(1024)

	fd := 1
	assert.Nil(t, thread.GetConnection(fd))
	thread.RegisterConnection(&Connection{fd: fd})

	conn := thread.GetConnection(fd)
	assert.NotNil(t, conn)
	assert.Equal(t, fd, conn.fd)
}

func TestThread_UnregisterConnection(t *testing.T) {
	thread := NewThread(1024)
	fd := 1
	thread.RegisterConnection(&Connection{fd: fd})

	conn := thread.GetConnection(fd)
	assert.NotNil(t, conn)
	assert.Equal(t, fd, conn.fd)

	// unregister connection
	thread.UnregisterConnection(conn)
	assert.Nil(t, thread.GetConnection(fd))
}

func TestThread_OfferAndPolling(t *testing.T) {
	thread := NewThread(1024)

	stop := make(chan struct{})
	go thread.Polling(stop, func(active int) {
		fmt.Println("active: ", active)
	})

	thread.Offer(1, 2, 3)

	go func() {
		time.Sleep(time.Second * 3)
		close(stop)
	}()
	<-stop
}