package transport

import (
	"net"
	"sync"
	"time"
)

// UDPTransport implements the Transport interface for UDP connections
type UDPTransport struct {
	conn     *net.UDPConn
	addr     *net.UDPAddr
	closed   bool
	mutex    sync.RWMutex
	timeout  time.Duration
}

// NewUDPTransport creates a new UDP transport
func NewUDPTransport() *UDPTransport {
	return &UDPTransport{
		timeout: time.Second * 30, // Default timeout
	}
}

// Listen starts listening on the specified address
func (u *UDPTransport) Listen(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	
	u.mutex.Lock()
	defer u.mutex.Unlock()
	
	u.conn = conn
	u.addr = udpAddr
	u.closed = false
	
	return nil
}

// Accept accepts a new connection (for UDP this is just returning the connection)
func (u *UDPTransport) Accept() (net.Conn, error) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	
	if u.conn == nil || u.closed {
		return nil, net.ErrClosed
	}
	
	return u.conn, nil
}

// Dial connects to the specified address
func (u *UDPTransport) Dial(addr string, timeout time.Duration) (net.Conn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}
	
	u.mutex.Lock()
	defer u.mutex.Unlock()
	
	u.conn = conn
	u.addr = udpAddr
	u.closed = false
	u.timeout = timeout
	
	return conn, nil
}

// Close closes the transport
func (u *UDPTransport) Close() error {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	
	if u.conn != nil {
		u.closed = true
		return u.conn.Close()
	}
	
	return nil
}

// IsClosed returns whether the transport is closed
func (u *UDPTransport) IsClosed() bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	
	return u.closed
}

// Send sends data to the specified address
func (u *UDPTransport) Send(data []byte, addr *net.UDPAddr) error {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	
	if u.conn == nil || u.closed {
		return net.ErrClosed
	}
	
	_, err := u.conn.WriteToUDP(data, addr)
	return err
}

// Receive receives data from the UDP connection
func (u *UDPTransport) Receive(buffer []byte) (*net.UDPAddr, int, error) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	if u.conn == nil || u.closed {
		return nil, 0, net.ErrClosed
	}

	u.conn.SetReadDeadline(time.Now().Add(u.timeout))
	n, addr, err := u.conn.ReadFromUDP(buffer)
	return addr, n, err
}

// SetTimeout sets the read/write timeout
func (u *UDPTransport) SetTimeout(timeout time.Duration) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	
	u.timeout = timeout
}

// GetLocalAddr returns the local address of the UDP connection
func (u *UDPTransport) GetLocalAddr() *net.UDPAddr {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	
	return u.addr
}