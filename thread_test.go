package znet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestThread(t *testing.T) {
	instance := NewThread(defaultThreadOptions())
	assert.NotNil(t, instance)
}

func TestThread_UseAndHandleRequest(t *testing.T) {
	
}
