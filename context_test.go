package znet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContext_Request(t *testing.T) {

}

func TestContext_Conn(t *testing.T) {
	ctx := &Context{conn: NewConnection(nil, 1)}
	assert.Equal(t, 1, ctx.Conn().fd)
}

func TestContext_Next(t *testing.T) {

}

func TestContext_Abort(t *testing.T) {

}

func TestContext_reset(t *testing.T) {

}
