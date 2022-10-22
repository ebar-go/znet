package znet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRouter(t *testing.T) {
	instance := NewRouter()
	assert.NotNil(t, instance)
}

func TestStandardHandler(t *testing.T) {
	type Request struct{}
	type Response struct{}
	handler := StandardHandler[Request, Response](func(ctx *Context, req *Request) (*Response, error) {
		return &Response{}, nil
	})

	assert.NotNil(t, handler)

}
