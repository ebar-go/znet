package poller

import (
	"net"
	"sync"
)

type Poller interface {
	Add(fd int) error
	Remove(fd int) error
	Wait() ([]int, error)
	SocketFD(conn net.Conn) int
}

var pollerInstance struct {
	once   sync.Once
	poller Poller
}

func Initialize(poller Poller) {
	pollerInstance.once.Do(func() {
		pollerInstance.poller = poller
	})
}

func Get() Poller {
	return pollerInstance.poller
}
