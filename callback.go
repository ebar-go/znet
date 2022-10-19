package znet

// Callback manage connection callback handlers.
type Callback struct {
	connect    ConnectionHandler
	disconnect ConnectionHandler
}

// OnConnect is called when the connection is established
func (callback *Callback) OnConnect(conn *Connection) {
	if callback.connect != nil {
		callback.connect(conn)
	}
}

// OnDisconnect is called when the connection is closed
func (callback *Callback) OnDisconnect(conn *Connection) {
	if callback.disconnect != nil {
		callback.disconnect(conn)
	}
}

// newCallback creates a new callback instance with the given handlers
func newCallback(connectHandler, disconnectHandler ConnectionHandler) *Callback {
	return &Callback{
		connect:    connectHandler,
		disconnect: disconnectHandler,
	}
}
