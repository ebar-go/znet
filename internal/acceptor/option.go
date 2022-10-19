package acceptor

import "runtime"

type Options struct {
	Core            int
	ReadBufferSize  int
	WriteBufferSize int
	Keepalive       bool
}

func DefaultOptions() *Options {
	return &Options{
		Core:            runtime.NumCPU(),
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		Keepalive:       true,
	}
}
