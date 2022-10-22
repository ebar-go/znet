package znet

import (
	"github.com/ebar-go/ego/utils/runtime/signal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMainReactor(t *testing.T) {
	reactor, err := NewMainReactor(defaultReactorOptions())

	assert.Nil(t, err)
	assert.NotNil(t, reactor)
}

func TestMainReactor_Run(t *testing.T) {
	reactor, err := NewMainReactor(defaultReactorOptions())

	assert.Nil(t, err)

	reactor.Run(signal.SetupSignalHandler())
}
