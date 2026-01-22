// Package interfaces defines the server-side interfaces for the video module
//
// The Server interface defines the methods that the server must implement to handle
// video streaming and management. These methods include starting and stopping the server,
// getting server status, stream information, client information, and adding/removing streams.
package interfaces

import (
	"context"
	"net/http"

	"yunyez/internal/video/types"
)

// Server defines the interface for an RTSP server
type Server interface {
	// Start starts the RTSP server
	Start(ctx context.Context) error
	
	// Stop stops the RTSP server
	Stop(ctx context.Context) error
	
	// GetStatus returns the current status of the server
	GetStatus() *types.ServerStats
	
	// GetStream returns information about a specific stream
	GetStream(streamID string) *types.Stream
	
	// GetStreams returns all active streams
	GetStreams() []*types.Stream
	
	// GetClient returns information about a specific client
	GetClient(sessionID string) *types.Client
	
	// GetClients returns all connected clients
	GetClients() []*types.Client
	
	// AddStream adds a new stream to the server
	AddStream(stream *types.Stream) error
	
	// RemoveStream removes a stream from the server
	RemoveStream(streamID string) error
	
	// GetHTTPHandler returns the HTTP handler for health checks and status
	GetHTTPHandler() http.Handler
}