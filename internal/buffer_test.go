package internal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewBuffer(t *testing.T) {
	buffer := NewBuffer[int](10)
	assert.NotNil(t, buffer)
}

func TestBuffer_Offer(t *testing.T) {
	buffer := NewBuffer[int](10)
	buffer.Offer(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
}

func TestBuffer_Polling(t *testing.T) {
	buffer := NewBuffer[int](10)
	stop := make(chan struct{})
	go buffer.Polling(stop, func(item int) {
		fmt.Println("item: ", item)
	})

	buffer.Offer(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	go func() {
		time.Sleep(time.Second * 3)
		close(stop)
	}()
	<-stop
}
