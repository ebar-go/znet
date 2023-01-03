package acceptor

import (
	"runtime"
	"time"
)

type Options struct {
	Core            int
	ReadBufferSize  int
	WriteBufferSize int
	Keepalive       bool
	WriteDeadline   time.Duration
	ReadDeadline    time.Duration
	LengthOffset    int
	ReuseThread     uint8
}

func DefaultOptions() Options {
	return Options{
		Core:            runtime.NumCPU(),
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		Keepalive:       false,
		WriteDeadline:   time.Second * 3,
		ReadDeadline:    time.Second * 3,
		LengthOffset:    4,
	}
}
