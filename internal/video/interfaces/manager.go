package interfaces

import (
	"net"

	"yunyez/internal/video/types"
)

// StreamManager manages media streams
type StreamManager interface {
	// AddStream adds a new stream
	AddStream(stream *types.Stream) error
	
	// RemoveStream removes a stream by ID
	RemoveStream(streamID string) error
	
	// GetStream retrieves a stream by ID
	GetStream(streamID string) *types.Stream
	
	// GetStreams retrieves all streams
	GetStreams() []*types.Stream
	
	// UpdateStream updates stream information
	UpdateStream(stream *types.Stream) error
	
	// GetStreamStats retrieves statistics for a stream
	GetStreamStats(streamID string) *types.StreamStats
}

// ConnectionManager manages client connections
type ConnectionManager interface {
	// AddClient adds a new client connection
	AddClient(client *types.Client) error
	
	// RemoveClient removes a client connection by session ID
	RemoveClient(sessionID string) error
	
	// GetClient retrieves a client by session ID
	GetClient(sessionID string) *types.Client
	
	// GetClientsByStream retrieves all clients connected to a specific stream
	GetClientsByStream(streamID string) []*types.Client
	
	// GetClients retrieves all connected clients
	GetClients() []*types.Client
	
	// UpdateClient updates client information
	UpdateClient(client *types.Client) error
	
	// HandleRTSPRequest handles an incoming RTSP request
	HandleRTSPRequest(conn net.Conn, req *types.RTSPRequest) (*types.RTSPResponse, error)
}