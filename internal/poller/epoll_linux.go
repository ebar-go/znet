//go:build linux

package poller

import (
	"golang.org/x/sys/unix"
	"net"
	"reflect"
	"sync"
	"syscall"
)

// Epoll implements of Poller for linux
type Epoll struct {
	lock sync.RWMutex
	// 注册的事件的文件描述符
	fd int
	// max event size, default: 100
	maxEventSize int

	connBuffers []int
	events      []unix.EpollEvent
}

func (e *Epoll) Add(fd int) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	// 向 epoll 实例注册文件描述符对应的事件
	// POLLIN(0x1) 表示对应的文件描述字可以读
	// POLLHUP(0x10) 表示对应的文件描述字被挂起
	// EPOLLET(0x80000000) 将EPOLL设为边缘触发(Edge Triggered)模式，这是相对于水平触发(Level Triggered)来说的。缺省是水平触发(Level Triggered)。

	// 只有当链接有数据可以读或者连接被关闭时，wait才会唤醒
	err := unix.EpollCtl(e.fd,
		unix.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP | unix.EPOLLET, Fd: int32(fd)})

	if err != nil {
		return err
	}
	return nil

}

func (e *Epoll) Remove(fd int) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	// 向 epoll 实例删除文件描述符对应的事件
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	return nil
}

func (e *Epoll) Wait() ([]int, error) {
	events := e.events
	var (
		n   int
		err error
	)
	for {
		n, err = unix.EpollWait(e.fd, events, 100)
		if err == nil {
			break
		}
		if err == unix.EINTR {
			continue
		}
		return nil, err
	}
	e.lock.RLock()

	connections := e.connBuffers[:0]
	for i := 0; i < n; i++ {
		connections = append(connections, int(e.events[i].Fd))
	}

	e.lock.RUnlock()
	return connections, nil
}

func (e *Epoll) Close() error {
	e.lock.Lock()
	defer e.lock.Unlock()
	return unix.Close(e.fd)
}

func NewPollerWithBuffer(size int) (Poller, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	return &Epoll{
		fd:           fd,
		maxEventSize: size,
		events:       make([]unix.EpollEvent, size, size),
		connBuffers:  make([]int, size, size),
	}, nil
}

// SocketFD get socket connection fd
func (e *Epoll) SocketFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
