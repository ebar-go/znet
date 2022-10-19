//go:build windows && cgo

package poller

import (
	"github.com/ebar-go/znet/internal/poller/wepoll"
)

type epoll = wepoll.Epoll

func NewPollerWithBuffer(size int) (Poller, error) {
	return wepoll.NewPollerWithBuffer(size)
}
