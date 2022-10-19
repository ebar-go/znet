package acceptor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	options := DefaultOptions()
	assert.NotNil(t, options)
}
