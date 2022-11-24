package poller

import (
	"net"
)

type Poller interface {
	Add(fd int) error
	Remove(fd int) error
	Wait() ([]int, error)
	SocketFD(conn net.Conn) int
}
