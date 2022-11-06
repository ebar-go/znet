package znet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOptions_NewMainReactor(t *testing.T) {
	reactor := defaultOptions().NewReactor()
	assert.NotNil(t, reactor)
}

func TestOptions_Validate(t *testing.T) {
	type fields struct {
		Debug        bool
		OnConnect    ConnectionHandler
		OnDisconnect ConnectionHandler
		Middlewares  []HandleFunc
		Reactor      ReactorOptions
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "default options",
			fields: fields{
				Debug:        false,
				OnConnect:    nil,
				OnDisconnect: nil,
				Middlewares:  nil,
				Reactor:      defaultReactorOptions(),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &Options{
				Debug:        tt.fields.Debug,
				OnConnect:    tt.fields.OnConnect,
				OnDisconnect: tt.fields.OnDisconnect,
				Middlewares:  tt.fields.Middlewares,
				Reactor:      tt.fields.Reactor,
			}
			assert.Equal(t, tt.wantErr, options.Validate())
		})
	}
}

func Test_defaultOptions(t *testing.T) {
	options := defaultOptions()
	assert.Equal(t, defaultReactorOptions(), options.Reactor)
}

func Test_defaultReactorOptions(t *testing.T) {
	options := defaultReactorOptions()
	assert.NotNil(t, options)
}
