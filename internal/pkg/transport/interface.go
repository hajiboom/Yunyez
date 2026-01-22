package transport

import (
	"net"
	"time"
)

// Transport defines the interface for different transport protocols
type Transport interface {
	// Listen starts listening on the specified address
	Listen(addr string) error
	
	// Accept accepts a new connection
	Accept() (net.Conn, error)
	
	// Dial connects to the specified address
	Dial(addr string, timeout time.Duration) (net.Conn, error)
	
	// Close closes the transport
	Close() error
	
	// IsClosed returns whether the transport is closed
	IsClosed() bool
}

// RTPTransport defines the interface for RTP transport
type RTPTransport interface {
	// Send sends an RTP packet
	Send(packet []byte) error
	
	// Receive receives an RTP packet
	Receive() ([]byte, error)
	
	// SetTimeout sets the read/write timeout
	SetTimeout(timeout time.Duration)
	
	// Close closes the RTP transport
	Close() error
}

// RTCPTransport defines the interface for RTCP transport
type RTCPTransport interface {
	// Send sends an RTCP packet
	Send(packet []byte) error
	
	// Receive receives an RTCP packet
	Receive() ([]byte, error)
	
	// SetTimeout sets the read/write timeout
	SetTimeout(timeout time.Duration)
	
	// Close closes the RTCP transport
	Close() error
}