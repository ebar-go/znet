package acceptor

import (
	"github.com/ebar-go/ego/utils/binary"
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
	Endian          binary.Endian
}

func DefaultOptions() *Options {
	return &Options{
		Core:            runtime.NumCPU(),
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		Keepalive:       false,
		WriteDeadline:   time.Second * 3,
		ReadDeadline:    time.Second * 3,
		Endian:          binary.BigEndian(),
	}
}
