package transport

import (
	"net"
	"time"
)

// TCPTransport implements the Transport interface for TCP connections
type TCPTransport struct {
	listener net.Listener
	conn     net.Conn
	addr     string
	closed   bool
}

// NewTCPTransport creates a new TCP transport
func NewTCPTransport() *TCPTransport {
	return &TCPTransport{}
}

// Listen starts listening on the specified address
func (t *TCPTransport) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	
	t.listener = listener
	t.addr = addr
	t.closed = false
	
	return nil
}

// Accept accepts a new connection
func (t *TCPTransport) Accept() (net.Conn, error) {
	if t.listener == nil {
		return nil, net.ErrClosed
	}
	
	conn, err := t.listener.Accept()
	if err != nil {
		return nil, err
	}
	
	t.conn = conn
	return conn, nil
}

// Dial connects to the specified address
func (t *TCPTransport) Dial(addr string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	
	t.conn = conn
	t.addr = addr
	t.closed = false
	
	return conn, nil
}

// Close closes the transport
func (t *TCPTransport) Close() error {
	t.closed = true
	
	var err error
	if t.conn != nil {
		err = t.conn.Close()
	}
	
	if t.listener != nil {
		listenerErr := t.listener.Close()
		if err == nil {
			err = listenerErr
		}
	}
	
	return err
}

// IsClosed returns whether the transport is closed
func (t *TCPTransport) IsClosed() bool {
	return t.closed
}