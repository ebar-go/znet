//go:build darwin || netbsd || freebsd || openbsd || dragonfly
// +build darwin netbsd freebsd openbsd dragonfly

package poller

import (
	"sync"
	"syscall"
)

type epoll struct {
	fd          int
	ts          syscall.Timespec
	changes     []syscall.Kevent_t
	mu          *sync.RWMutex
	connections []int
	events      []syscall.Kevent_t
}

func NewPollerWithBuffer(count int) (Poller, error) {
	p, err := syscall.Kqueue()
	if err != nil {
		panic(err)
	}
	_, err = syscall.Kevent(p, []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}, nil, nil)
	if err != nil {
		panic(err)
	}

	return &epoll{
		fd:          p,
		ts:          syscall.NsecToTimespec(1e9),
		mu:          &sync.RWMutex{},
		connections: make([]int, count, count),
		events:      make([]syscall.Kevent_t, count, count),
	}, nil
}

func (e *epoll) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return syscall.Close(e.fd)
}

func (e *epoll) Add(fd int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.changes = append(e.changes,
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD | syscall.EV_EOF, Filter: syscall.EVFILT_READ,
		},
	)

	return nil
}

func (e *epoll) Remove(fd int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.changes) <= 1 {
		e.changes = nil
	} else {
		changes := make([]syscall.Kevent_t, 0, len(e.changes)-1)
		ident := uint64(fd)
		for _, ke := range e.changes {
			if ke.Ident != ident {
				changes = append(changes, ke)
			}
		}
		e.changes = changes
	}

	return nil
}

func (e *epoll) Wait() ([]int, error) {
	e.mu.RLock()
	changes := e.changes
	e.mu.RUnlock()

retry:
	n, err := syscall.Kevent(e.fd, changes, e.events, &e.ts)
	if err != nil {
		if err == syscall.EINTR {
			goto retry
		}
		return nil, err
	}

	var connections = e.connections[:0]
	e.mu.RLock()
	for i := 0; i < n; i++ {
		connections = append(connections, int(e.events[i].Ident))
	}
	e.mu.RUnlock()
	return connections, nil
}
