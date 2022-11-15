package znet

// Callback manage connection callback handlers.
type Callback struct {
	open  ConnectionHandler
	close ConnectionHandler
}

// triggerOpenEvent is called when the connection is established
func (callback *Callback) triggerOpenEvent(conn *Connection) {
	if callback.open != nil {
		callback.open(conn)
	}
}

// triggerCloseEvent is called when the connection is closed
func (callback *Callback) triggerCloseEvent(conn *Connection) {
	if callback.close != nil {
		callback.close(conn)
	}
}

// newCallback creates a new callback instance with the given handlers
func newCallback(open, close ConnectionHandler) *Callback {
	return &Callback{
		open:  open,
		close: close,
	}
}
