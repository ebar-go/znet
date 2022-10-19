//go:build windows && cgo

package wepoll

//#cgo windows LDFLAGS: -lws2_32 -lwsock32
//#include"wepoll.h"
import "C"
import (
	"errors"
	"net"
	"sync"
	"syscall"
)

type Epoll struct {
	fd          C.uintptr_t
	connections map[int]net.Conn
	lock        *sync.RWMutex
	buffer      []int
	events      []C.epoll_event
}

func NewPollerWithBuffer(count int) (*Epoll, error) {
	fd := C.epoll_create1(0)
	if fd == 0 {
		return nil, errors.New("epoll_create1 error")
	}
	return &Epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
		buffer:      make([]int, count, count),
		events:      make([]C.epoll_event, count, count),
	}, nil
}

func (e *Epoll) Close() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.connections = nil
	i := C.epoll_close(e.fd)
	if i == 0 {
		return nil
	} else {

		return errors.New(" an error occurred on epoll.close ")
	}
}

func (e *Epoll) SocketFD(conn net.Conn) int {
	fd := C.SOCKET(socketFDAsUint(conn))
	return int(fd)
}

func (e *Epoll) Add(fd int) error {
	// Extract file descriptor associated with the connection
	var ev C.epoll_event
	ev = C.set_epoll_event(C.EPOLLIN|C.EPOLLHUP, C.SOCKET(fd))
	e.lock.Lock()
	defer e.lock.Unlock()
	err := C.epoll_ctl(e.fd, C.EPOLL_CTL_ADD, C.SOCKET(fd), &ev)
	if err == -1 {
		return errors.New("C.EPOLL_CTL_ADD error ")
	}
	return nil
}

func (e *Epoll) Remove(fd int) error {

	var ev C.epoll_event
	err := C.epoll_ctl(e.fd, C.EPOLL_CTL_DEL, C.SOCKET(fd), &ev)
	if err == -1 {
		return errors.New("C.EPOLL_CTL_DEL error ")
	}
	return nil
}

func (e *Epoll) Wait() ([]int, error) {
	n := C.epoll_wait(e.fd, &e.events[0], 128, -1)
	if n == -1 {
		return nil, errors.New("Wait err")
	}

	var connections = e.buffer[:0]
	e.lock.RLock()
	for i := 0; i < int(n); i++ {
		fd := C.get_epoll_event(e.events[i])

		connections = append(connections, int(fd))
	}
	e.lock.RUnlock()

	return connections, nil
}

func socketFDAsUint(conn net.Conn) uint64 {
	if con, ok := conn.(syscall.Conn); ok {
		raw, err := con.SyscallConn()
		if err != nil {
			return 0
		}
		sfd := uint64(0)
		raw.Control(func(fd uintptr) {
			sfd = uint64(fd)
		})
		return sfd
	}
	return 0
}
