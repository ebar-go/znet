package znet

import "github.com/ebar-go/ego/component"

const (
	BeforeServerStart    = "beforeServerStart"
	AfterServerStart     = "afterServerStart"
	BeforeServerShutdown = "beforeServerShutdown"
	AfterServerShutdown  = "afterServerShutdown"
)

// RegisterEvent registers an event handler
func RegisterEvent(event string, callback func()) {
	component.Event().Listen(event, func(param any) {
		callback()
	})
}
