package poller

import (
	"net"
	"syscall"
)

type Poller interface {
	Add(fd int) error
	Remove(fd int) error
	Wait() ([]int, error)
}

// SocketFD get socket connection fd
func SocketFD(conn net.Conn) int {
	if con, ok := conn.(syscall.Conn); ok {
		raw, err := con.SyscallConn()
		if err != nil {
			return 0
		}
		sfd := 0
		_ = raw.Control(func(fd uintptr) {
			sfd = int(fd)
		})
		return sfd
	}
	return 0
}
