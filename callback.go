package znet

// Callback manage connection callback handlers.
type Callback struct {
	openHandler  ConnectionHandler
	closeHandler ConnectionHandler
	errorHandler func(ctx *Context, err error)
}

// onOpen is called when the connection is established
func (callback *Callback) onOpen(conn *Connection) {
	if callback.openHandler != nil {
		callback.openHandler(conn)
	}
}

// onClose is called when the connection is closed
func (callback *Callback) onClose(conn *Connection) {
	if callback.closeHandler != nil {
		callback.closeHandler(conn)
	}
}

func (callback *Callback) onError(ctx *Context, err error) {
	if callback.errorHandler != nil {
		callback.errorHandler(ctx, err)
	}
}
