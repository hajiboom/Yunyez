package video

import (
	"fmt"
	"sync"

	"yunyez/internal/video/types"
)

// StreamManagerImpl implements the StreamManager interface
type StreamManagerImpl struct {
	streams map[string]*types.Stream
	mutex   sync.RWMutex
}

// NewStreamManager creates a new stream manager
func NewStreamManager() *StreamManagerImpl {
	return &StreamManagerImpl{
		streams: make(map[string]*types.Stream),
	}
}

// AddStream adds a new stream
func (sm *StreamManagerImpl) AddStream(stream *types.Stream) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.streams[stream.ID]; exists {
		return fmt.Errorf("stream with ID %s already exists", stream.ID)
	}

	sm.streams[stream.ID] = stream
	return nil
}

// RemoveStream removes a stream by ID
func (sm *StreamManagerImpl) RemoveStream(streamID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.streams[streamID]; !exists {
		return fmt.Errorf("stream with ID %s does not exist", streamID)
	}

	delete(sm.streams, streamID)
	return nil
}

// GetStream retrieves a stream by ID
func (sm *StreamManagerImpl) GetStream(streamID string) *types.Stream {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stream, exists := sm.streams[streamID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	streamCopy := *stream
	return &streamCopy
}

// GetStreams retrieves all streams
func (sm *StreamManagerImpl) GetStreams() []*types.Stream {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var streams []*types.Stream
	for _, stream := range sm.streams {
		// Return copies to prevent external modification
		streamCopy := *stream
		streams = append(streams, &streamCopy)
	}

	return streams
}

// UpdateStream updates stream information
func (sm *StreamManagerImpl) UpdateStream(stream *types.Stream) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.streams[stream.ID]; !exists {
		return fmt.Errorf("stream with ID %s does not exist", stream.ID)
	}

	sm.streams[stream.ID] = stream
	return nil
}

// GetStreamStats retrieves statistics for a stream
func (sm *StreamManagerImpl) GetStreamStats(streamID string) *types.StreamStats {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stream, exists := sm.streams[streamID]
	if !exists {
		return nil
	}

	// In a real implementation, this would calculate actual stats
	// For now, we'll return basic info
	return &types.StreamStats{
		StreamID:         stream.ID,
		TotalConnections: 0, // Would be calculated from connection manager
		ActiveClients:    0, // Would be calculated from connection manager
		BytesSent:        0, // Would be tracked separately
		BytesReceived:    0, // Would be tracked separately
		BitrateKbps:      stream.Bitrate,
		StartTime:        stream.CreatedAt,
		Uptime:           0, // Would be calculated
	}
}