//go:build windows && cgo

package poller

import (
	"github.com/ebar-go/znet/internal/poller/wepoll"
)

func NewPollerWithBuffer(size int) (Poller, error) {
	return wepoll.NewPollerWithBuffer(size)
}
