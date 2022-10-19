package znet

type Connection struct {
	fd int
}

func (c *Connection) readLine(p []byte, offset int) (int, error) {
	return 0, nil
}
