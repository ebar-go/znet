package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSyncPoolProvider(t *testing.T) {
	provider := NewSyncPoolProvider[[]byte](func() interface{} {
		return make([]byte, 4096)
	})
	assert.NotNil(t, provider)

}

func TestSyncPoolProvider_Acquire(t *testing.T) {
	provider := NewSyncPoolProvider[[]byte](func() interface{} {
		return make([]byte, 4096)
	})

	bytes := provider.Acquire()
	assert.Len(t, bytes, 4096)
}

func TestSyncPoolProvider_Release(t *testing.T) {
	provider := NewSyncPoolProvider[[]byte](func() interface{} {
		return make([]byte, 4096)
	})

	bytes := provider.Acquire()
	assert.Len(t, bytes, 4096)

	copy(bytes[:3], []byte("foo"))
	assert.Equal(t, []byte("foo"), bytes[:3])

	provider.Release(bytes)
}
