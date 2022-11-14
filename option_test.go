package znet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOptions_NewMainReactor(t *testing.T) {
	reactor := defaultOptions().NewReactorOrDie()
	assert.NotNil(t, reactor)
}

func TestOptions_Validate(t *testing.T) {

}

func Test_defaultOptions(t *testing.T) {
	options := defaultOptions()
	assert.Equal(t, defaultReactorOptions(), options.Reactor)
}

func Test_defaultReactorOptions(t *testing.T) {
	options := defaultReactorOptions()
	assert.NotNil(t, options)
}
