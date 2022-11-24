//go:build windows && cgo

package poller

import (
	"github.com/ebar-go/znet/poller/wepoll"
)

func NewPollerWithBuffer(size int) (Poller, error) {
	return wepoll.NewPollerWithBuffer(size)
}
