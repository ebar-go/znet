package acceptor

type Options struct {
	core            int
	readBufferSize  int
	writeBufferSize int
	keepalive       bool
}
