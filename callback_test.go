package znet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCallback_OnConnect(t *testing.T) {
	connected := 0
	callback := newCallback(func(conn *Connection) {
		connected++
	}, nil)

	assert.NotNil(t, callback)
	assert.Equal(t, connected, 0)

	callback.OnOpen(&Connection{})
	assert.Equal(t, connected, 1)
}

func TestCallback_OnDisconnect(t *testing.T) {
	connected := 0
	callback := newCallback(func(conn *Connection) {
		connected++
	}, func(conn *Connection) {
		connected--
	})

	assert.NotNil(t, callback)
	assert.Equal(t, connected, 0)

	callback.OnOpen(&Connection{})
	assert.Equal(t, connected, 1)

	callback.OnClose(&Connection{})
	assert.Equal(t, connected, 0)
}
